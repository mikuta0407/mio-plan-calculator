let wasmReady = false;

const go = new Go();
WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
    .then((res) => {
        go.run(res.instance);
        wasmReady = true;
        document.getElementById("calcBtn").disabled = false;
        document.getElementById("calcBtn").textContent = "最安値を計算";
        document.getElementById("rangeBtn").disabled = false;
        document.getElementById("rangeBtn").textContent = "1GB単位で一覧表示";
        document.getElementById("loading").textContent = "";
    })
    .catch((err) => {
        document.getElementById("loading").textContent =
            "WASMの読み込みに失敗しました: " + err;
    });

// Read per-type constraints from the form.
// Returns {voiceMin, voiceMax, smsMin, smsMax, esimMin, esimMax, dataMin, dataMax}.
function getConstraints() {
    const iv = (id) => {
        const v = document.getElementById(id).value;
        return v === "" ? -1 : parseInt(v, 10);
    };
    return {
        voiceMin: iv("voice-min"),
        voiceMax: iv("voice-max"),
        smsMin:   iv("sms-min"),
        smsMax:   iv("sms-max"),
        esimMin:  iv("esim-min"),
        esimMax:  iv("esim-max"),
        dataMin:  iv("data-min"),
        dataMax:  iv("data-max"),
    };
}

function callFindCheapest(c, minGB, maxGB) {
    return findCheapest(
        c.voiceMin, c.voiceMax,
        c.smsMin,   c.smsMax,
        c.esimMin,  c.esimMax,
        c.dataMin,  c.dataMax,
        minGB, maxGB,
    );
}

function comboCompactLabel(combo) {
    return combo.types.map((t) => `${t.name}: ${t.label}`).join(" / ");
}

function buildResultHTML(res) {
    const fmt = (n) => n.toLocaleString("ja-JP");
    const discountLabel = res.discount > 0 ? `割引（${res.lines}回線 × ¥100）` : "割引（複数回線割引）";
    const discountValue = res.discount > 0 ? `－¥${fmt(res.discount)}` : '<span style="font-size:0.9rem;color:#718096">対象外</span>';
    let html = `
      <h2>計算結果</h2>
      <div class="result-summary">
        <div class="stat">
          <div class="label">プラン合計（割引前）</div>
          <div class="value">¥${fmt(res.bestCost)}<span style="font-size:0.8rem;font-weight:400">/月</span></div>
        </div>
        <div class="stat">
          <div class="label">${discountLabel}</div>
          <div class="value">${discountValue}</div>
        </div>
        <div class="stat highlight">
          <div class="label">最安値（割引後）</div>
          <div class="value">¥${fmt(res.finalCost)}<span style="font-size:0.8rem;font-weight:400">/月</span></div>
        </div>
      </div>
      <h2>最安の組み合わせ（${res.combos.length}パターン）</h2>
      <ul class="combo-list">`;
    for (const combo of res.combos) {
        html += `<li>
          <div class="combo-header">
            <span class="gb">合計 ${combo.totalGB}GB</span>
            <span class="per-gb">¥${combo.pricePerGB.toFixed(1)}/GB</span>
          </div>
          <div class="type-breakdown">`;
        for (const t of combo.types) {
            html += `<div class="type-line"><span class="type-name">${t.name}(${t.lines}回線)</span>${t.label} [${t.totalGB}GB]</div>`;
        }
        html += `</div></li>`;
    }
    html += `</ul>`;
    return html;
}

function buildRangeHTML(minGB, maxGB, getFn, canvasId) {
    const fmt = (n) => n.toLocaleString("ja-JP");

    // First pass: collect all results
    const results = [];
    for (let gb = minGB; gb <= maxGB; gb++) {
        results.push({ gb, res: getFn(gb) });
    }

    // Compute per-GB price stats for color highlights
    const validPPG = results
        .filter((r) => r.res.found)
        .map((r) => r.res.finalCost / r.gb);
    const minPPG = validPPG.length ? Math.min(...validPPG) : null;
    const maxPPG = validPPG.length ? Math.max(...validPPG) : null;
    const sorted = [...validPPG].sort((a, b) => a - b);
    const medianPPG = sorted.length
        ? sorted[Math.floor(sorted.length / 2)]
        : null;
    // Find the value closest to median (excludes min/max)
    const medianTarget = sorted.filter((v) => v !== minPPG && v !== maxPPG);
    const closestToMedian = medianTarget.length
        ? medianTarget.reduce((a, b) =>
              Math.abs(a - medianPPG) <= Math.abs(b - medianPPG) ? a : b)
        : null;

    const data = [];
    let html = `<h2>1GB単位の一覧（${minGB}GB〜${maxGB}GB）</h2>
      <div class="range-legend">
        <span class="legend-item legend-best">単価 最安</span>
        <span class="legend-item legend-median">単価 中央値</span>
        <span class="legend-item legend-worst">単価 最高</span>
      </div>
      <div class="range-scroll"><table class="range-table">
      <thead><tr><th>合計GB</th><th>回線数</th><th>最安値(割引後)</th><th>単価(/GB)</th><th>組み合わせ</th></tr></thead>
      <tbody>`;

    for (const { gb, res } of results) {
        if (!res.found) {
            data.push(null);
            html += `<tr class="no-result"><td>${gb}GB</td><td colspan="4">組み合わせなし</td></tr>`;
        } else {
            const perGB = res.finalCost / gb;
            data.push({ gb, price: res.finalCost, pricePerGB: perGB });
            const labels = res.combos.map((c) => comboCompactLabel(c)).join("<br>");
            let rowClass = "";
            if (perGB === minPPG) rowClass = "row-best";
            else if (perGB === maxPPG) rowClass = "row-worst";
            else if (closestToMedian !== null && perGB === closestToMedian) rowClass = "row-median";
            html += `<tr${rowClass ? ` class="${rowClass}"` : ""}>
              <td><strong>${gb}GB</strong></td>
              <td>${res.lines}</td>
              <td>¥${fmt(res.finalCost)}</td>
              <td>¥${perGB.toFixed(1)}</td>
              <td class="combos">${labels}</td>
            </tr>`;
        }
    }
    html += `</tbody></table></div>
      <div class="chart-wrap">
        <canvas id="${canvasId}" height="260"></canvas>
        <div class="chart-tooltip" id="${canvasId}-tip"></div>
      </div>
      <div class="chart-wrap">
        <canvas id="${canvasId}-pgb" height="260"></canvas>
        <div class="chart-tooltip" id="${canvasId}-pgb-tip"></div>
      </div>`;
    return { html, data };
}

function drawChart(canvasId, data, title) {
    const canvas = document.getElementById(canvasId);
    if (!canvas) return;
    const W = canvas.offsetWidth || 640;
    const H = 260;
    const dpr = window.devicePixelRatio || 1;
    canvas.width = W * dpr;
    canvas.height = H * dpr;
    const ctx = canvas.getContext("2d");
    ctx.scale(dpr, dpr);

    const PAD = { top: 36, right: 24, bottom: 44, left: 68 };
    const cw = W - PAD.left - PAD.right;
    const ch = H - PAD.top - PAD.bottom;

    const points = data.filter((d) => d !== null);
    if (points.length === 0) return;

    if (title) {
        ctx.font = `bold 12px -apple-system, sans-serif`;
        ctx.fillStyle = "#2d3748";
        ctx.textAlign = "center";
        ctx.fillText(title, W / 2, 16);
    }

    const maxPrice = Math.max(...points.map((d) => d.price));
    const minPrice = Math.min(...points.map((d) => d.price));
    const priceRange = maxPrice - minPrice || 1;
    const yMax = maxPrice + priceRange * 0.1;
    const yMin = Math.max(0, minPrice - priceRange * 0.1);

    const toX = (i) => PAD.left + (i / (data.length - 1)) * cw;
    const toY = (price) => PAD.top + ch - ((price - yMin) / (yMax - yMin)) * ch;

    // Grid lines & y-axis labels
    ctx.font = `11px -apple-system, sans-serif`;
    ctx.fillStyle = "#718096";
    const yTicks = 5;
    for (let t = 0; t <= yTicks; t++) {
        const price = yMin + (yMax - yMin) * (t / yTicks);
        const y = toY(price);
        ctx.strokeStyle = "#e2e8f0";
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(PAD.left, y);
        ctx.lineTo(PAD.left + cw, y);
        ctx.stroke();
        ctx.textAlign = "right";
        ctx.fillText("¥" + Math.round(price).toLocaleString("ja-JP"), PAD.left - 6, y + 4);
    }

    // X-axis labels
    const step = Math.ceil(data.length / 10);
    ctx.textAlign = "center";
    for (let i = 0; i < data.length; i += step) {
        const d = data[i];
        if (!d) continue;
        ctx.fillStyle = "#718096";
        ctx.fillText(d.gb + "GB", toX(i), H - PAD.bottom + 16);
    }

    // Area fill
    const grad = ctx.createLinearGradient(0, PAD.top, 0, PAD.top + ch);
    grad.addColorStop(0, "rgba(66,153,225,0.25)");
    grad.addColorStop(1, "rgba(66,153,225,0)");
    ctx.beginPath();
    let started = false;
    for (let i = 0; i < data.length; i++) {
        if (!data[i]) { started = false; continue; }
        const x = toX(i), y = toY(data[i].price);
        if (!started) {
            ctx.moveTo(x, toY(yMin));
            ctx.lineTo(x, y);
            started = true;
        } else {
            ctx.lineTo(x, y);
        }
    }
    for (let i = data.length - 1; i >= 0; i--) {
        if (data[i]) { ctx.lineTo(toX(i), toY(yMin)); break; }
    }
    ctx.fillStyle = grad;
    ctx.fill();

    // Line
    ctx.strokeStyle = "#4299e1";
    ctx.lineWidth = 2;
    ctx.lineJoin = "round";
    started = false;
    ctx.beginPath();
    for (let i = 0; i < data.length; i++) {
        if (!data[i]) { started = false; continue; }
        const x = toX(i), y = toY(data[i].price);
        if (!started) { ctx.moveTo(x, y); started = true; }
        else ctx.lineTo(x, y);
    }
    ctx.stroke();

    // Dots
    for (let i = 0; i < data.length; i++) {
        if (!data[i]) continue;
        ctx.beginPath();
        ctx.arc(toX(i), toY(data[i].price), data.length > 30 ? 2 : 3.5, 0, Math.PI * 2);
        ctx.fillStyle = "#2b6cb0";
        ctx.fill();
    }

    // Hover tooltip
    const tip = document.getElementById(canvasId + "-tip");
    canvas.onmousemove = (e) => {
        const rect = canvas.getBoundingClientRect();
        const mx = e.clientX - rect.left;
        const my = e.clientY - rect.top;
        let closest = null, minDist = Infinity;
        for (let i = 0; i < data.length; i++) {
            if (!data[i]) continue;
            const dx = toX(i) - mx, dy = toY(data[i].price) - my;
            const dist = Math.sqrt(dx * dx + dy * dy);
            if (dist < minDist) { minDist = dist; closest = { i, d: data[i] }; }
        }
        if (closest && minDist < 30) {
            tip.style.display = "block";
            tip.textContent = `${closest.d.gb}GB → ¥${closest.d.price.toLocaleString("ja-JP")}/月`;
            tip.style.left = toX(closest.i) + 8 + "px";
            tip.style.top = toY(closest.d.price) - 28 + "px";
        } else {
            tip.style.display = "none";
        }
    };
    canvas.onmouseleave = () => { tip.style.display = "none"; };
}

function calculate() {
    const minGB = parseInt(document.getElementById("minGB").value, 10);
    const maxGB = parseInt(document.getElementById("maxGB").value, 10);
    const div = document.getElementById("result");
    div.style.display = "block";
    if (isNaN(minGB) || isNaN(maxGB) || minGB < 1 || maxGB < minGB) {
        div.innerHTML = `<div class="error">⚠️ 入力値が不正です。</div>`;
        return;
    }
    const c = getConstraints();
    const res = callFindCheapest(c, minGB, maxGB);
    if (!res.found) {
        div.innerHTML = `<div class="error">⚠️ ${res.message}</div>`;
        return;
    }
    div.innerHTML = buildResultHTML(res);
}

function calculateRange() {
    const minGB = parseInt(document.getElementById("minGB").value, 10);
    const maxGB = parseInt(document.getElementById("maxGB").value, 10);
    const div = document.getElementById("range-result");
    div.style.display = "block";
    if (isNaN(minGB) || isNaN(maxGB) || minGB < 1 || maxGB < minGB) {
        div.innerHTML = `<div class="error">⚠️ 入力値が不正です。</div>`;
        return;
    }
    const c = getConstraints();
    const { html, data } = buildRangeHTML(
        minGB, maxGB,
        (gb) => callFindCheapest(c, gb, gb),
        "chart-fixed",
    );
    div.innerHTML = html;
    requestAnimationFrame(() => {
        drawChart("chart-fixed", data, "月額最安値（割引後）");
        drawChart(
            "chart-fixed-pgb",
            data.map((d) => d ? { gb: d.gb, price: d.pricePerGB } : null),
            "GB単価（¥/GB）",
        );
    });
}
