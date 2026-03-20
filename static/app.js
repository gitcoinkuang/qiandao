let state = {
    tasks: [],
    history: [],
    editingId: null,
    taskFilter: "all",
    taskSearch: "",
};

const TEXT = {
    requestFailed: "\u8bf7\u6c42\u5931\u8d25",
    unexpectedError: "\u53d1\u751f\u4e86\u672a\u9884\u671f\u7684\u9519\u8bef",
    totalTasks: "\u4efb\u52a1\u603b\u6570",
    enabledTasks: "\u542f\u7528\u4efb\u52a1",
    recentSuccess: "\u6700\u8fd1\u6210\u529f",
    averageDuration: "\u5e73\u5747\u8017\u65f6",
    totalTasksNote: "\u6240\u6709\u4efb\u52a1",
    enabledTasksNote: "\u5f53\u524d\u4f1a\u53c2\u4e0e\u8fd0\u884c",
    recentSuccessNote: "\u6700\u8fd1\u4e00\u6bb5\u65f6\u95f4\u6210\u529f\u6b21\u6570",
    averageDurationNote: "\u8fd1\u671f\u4efb\u52a1\u5e73\u5747\u54cd\u5e94",
    noTasks: "\u8fd8\u6ca1\u6709\u4efb\u52a1\u3002\u4f60\u53ef\u4ee5\u4ece\u53f3\u4fa7\u76f4\u63a5\u521b\u5efa\u7b2c\u4e00\u4e2a\u4efb\u52a1\uff0c\u6216\u8005\u5148\u7c98\u8d34\u4e00\u4e2a Curl \u547d\u4ee4\u505a\u89e3\u6790\u3002",
    noHistory: "\u8fd8\u6ca1\u6709\u6267\u884c\u5386\u53f2\u3002\u8fd0\u884c\u4e00\u6b21\u4efb\u52a1\u4e4b\u540e\uff0c\u8fd9\u91cc\u4f1a\u5f00\u59cb\u5c55\u793a\u5b8c\u6574\u65f6\u95f4\u7ebf\u3002",
    enabled: "\u5df2\u542f\u7528",
    disabled: "\u5df2\u505c\u7528",
    globalSchedule: "\u8ddf\u968f\u5168\u5c40\u8c03\u5ea6",
    timeout: "\u8d85\u65f6",
    retry: "\u91cd\u8bd5",
    lastRun: "\u6700\u8fd1\u8fd0\u884c",
    neverRun: "\u4ece\u672a\u6267\u884c",
    lastDuration: "\u6700\u8fd1\u8017\u65f6",
    run: "\u8fd0\u884c",
    edit: "\u7f16\u8f91",
    del: "\u5220\u9664",
    statusCode: "\u72b6\u6001\u7801",
    duration: "\u8017\u65f6",
    parseSuccess: "Curl \u89e3\u6790\u6210\u529f",
    taskCreated: "\u4efb\u52a1\u5df2\u521b\u5efa",
    taskUpdated: "\u4efb\u52a1\u5df2\u66f4\u65b0",
    taskDeleted: "\u4efb\u52a1\u5df2\u5220\u9664",
    taskExecuted: "\u4efb\u52a1\u5df2\u6267\u884c",
    allExecuted: "\u5168\u90e8\u542f\u7528\u4efb\u52a1\u5df2\u6267\u884c",
    historyCleared: "\u5386\u53f2\u8bb0\u5f55\u5df2\u6e05\u7a7a",
    scheduleSaved: "\u5b9a\u65f6\u8bbe\u7f6e\u5df2\u4fdd\u5b58",
    scheduleChecked: "\u5b9a\u65f6\u68c0\u67e5\u5df2\u6267\u884c",
    notifySaved: "\u901a\u77e5\u8bbe\u7f6e\u5df2\u4fdd\u5b58",
    notifyTestSent: "\u6d4b\u8bd5\u901a\u77e5\u5df2\u53d1\u9001",
    securitySaved: "\u5b89\u5168\u8bbe\u7f6e\u5df2\u4fdd\u5b58",
    formReset: "\u8868\u5355\u5df2\u91cd\u7f6e",
    createTask: "\u521b\u5efa\u4efb\u52a1",
    editTaskPrefix: "\u7f16\u8f91\u4efb\u52a1 #",
    taskPreviewHint: "\u8fd9\u91cc\u4f1a\u663e\u793a\u89e3\u6790\u7ed3\u679c\u6216\u5f53\u524d\u4efb\u52a1\u7684\u8be6\u7ec6\u4fe1\u606f\u3002",
    confirmDelete: "\u786e\u5b9a\u5220\u9664\u8fd9\u4e2a\u4efb\u52a1\u5417\uff1f",
    confirmClear: "\u786e\u5b9a\u6e05\u7a7a\u6240\u6709\u5386\u53f2\u8bb0\u5f55\u5417\uff1f",
    statusSuccess: "\u6210\u529f",
    statusFailed: "\u5931\u8d25",
    statusIdle: "\u7a7a\u95f2",
    healthSuccess: "\u6210\u529f\u8fd0\u884c",
    healthFailed: "\u5931\u8d25\u8fd0\u884c",
    healthAvg: "\u5e73\u5747\u8017\u65f6",
    healthNoData: "\u6682\u65e0\u8fd0\u884c\u6570\u636e",
    recentFeedEmpty: "\u6682\u65e0\u8fd1\u671f\u6d3b\u52a8",
    filterAll: "\u5168\u90e8",
    filterSuccess: "\u6210\u529f",
    filterFailed: "\u5931\u8d25",
    filterIdle: "\u7a7a\u95f2",
    taskSchedulePrefix: "\u5355\u4efb\u52a1 ",
};

const $ = (id) => document.getElementById(id);
const messageEl = $("message");

async function api(url, options = {}) {
    const response = await fetch(url, {
        headers: { "Content-Type": "application/json", ...(options.headers || {}) },
        ...options,
    });
    const data = await response.json();
    if (!response.ok || data.success === false) {
        throw new Error(data.error || TEXT.requestFailed);
    }
    return data;
}

function showMessage(text, type = "success") {
    messageEl.textContent = text;
    messageEl.className = `message ${type}`;
    window.clearTimeout(showMessage.timer);
    showMessage.timer = window.setTimeout(() => {
        messageEl.className = "message hidden";
    }, 3200);
}

function escapeHTML(value) {
    return String(value ?? "")
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#39;");
}

function pad(value) {
    return String(value).padStart(2, "0");
}

function statusClass(status) {
    if (status === "success") return "status-success";
    if (status === "failed") return "status-failed";
    return "status-idle";
}

function statusLabel(status) {
    if (status === "success") return TEXT.statusSuccess;
    if (status === "failed") return TEXT.statusFailed;
    return TEXT.statusIdle;
}

function renderSummary(stats) {
    const cards = [
        [TEXT.totalTasks, stats.total_tasks, TEXT.totalTasksNote],
        [TEXT.enabledTasks, stats.enabled_tasks, TEXT.enabledTasksNote],
        [TEXT.recentSuccess, stats.recent_success, TEXT.recentSuccessNote],
        [TEXT.averageDuration, `${stats.avg_duration_ms} ms`, TEXT.averageDurationNote],
    ];

    $("summaryGrid").innerHTML = cards.map(([label, value, note]) => `
        <div class="summary-card">
            <span>${label}</span>
            <strong>${value}</strong>
            <p>${note}</p>
        </div>
    `).join("");
}

function getFilteredTasks(tasks) {
    return tasks.filter((task) => {
        const search = state.taskSearch.trim().toLowerCase();
        const matchesSearch = !search || task.name.toLowerCase().includes(search) || task.url.toLowerCase().includes(search);
        const matchesFilter = state.taskFilter === "all" || (task.last_status || "idle") === state.taskFilter;
        return matchesSearch && matchesFilter;
    });
}

function renderTasks(tasks) {
    const filtered = getFilteredTasks(tasks);
    const target = $("taskList");
    if (!filtered.length) {
        target.innerHTML = `<div class="task-card"><div class="meta">${TEXT.noTasks}</div></div>`;
        return;
    }

    target.innerHTML = filtered.map((task) => `
        <div class="task-card">
            <div class="task-top">
                <div>
                    <div class="task-title">${escapeHTML(task.name)}</div>
                    <div class="meta">${escapeHTML(task.method)} · ${escapeHTML(task.url)}</div>
                </div>
                <span class="status-badge ${statusClass(task.last_status)}">${statusLabel(task.last_status)}</span>
            </div>

            <div class="pill-row">
                <span class="pill">${task.enabled ? TEXT.enabled : TEXT.disabled}</span>
                <span class="pill">${task.schedule_enabled ? `${TEXT.taskSchedulePrefix}${pad(task.schedule_hour)}:${pad(task.schedule_minute)}` : TEXT.globalSchedule}</span>
                <span class="pill">${TEXT.timeout} ${task.timeout_seconds}s</span>
                <span class="pill">${TEXT.retry} ${task.retry_count}</span>
            </div>

            <div class="meta">
                ${TEXT.lastRun}：${task.last_run_at || TEXT.neverRun}<br>
                ${TEXT.lastDuration}：${task.last_duration_ms || 0} ms
            </div>

            <div class="task-actions">
                <button class="primary" onclick="runTask(${task.id})">${TEXT.run}</button>
                <button class="secondary" onclick="editTask(${task.id})">${TEXT.edit}</button>
                <button class="danger" onclick="deleteTask(${task.id})">${TEXT.del}</button>
            </div>
        </div>
    `).join("");
}

function renderHistory(history) {
    const target = $("historyList");
    if (!history.length) {
        target.innerHTML = `<div class="history-card"><div class="meta">${TEXT.noHistory}</div></div>`;
        return;
    }

    target.innerHTML = history.map((item) => `
        <div class="history-card">
            <div class="task-top">
                <div>
                    <div class="task-title">${escapeHTML(item.task_name)}</div>
                    <div class="meta">${escapeHTML(item.triggered_by)} · ${escapeHTML(item.created_at)}</div>
                </div>
                <span class="status-badge ${statusClass(item.status)}">${statusLabel(item.status)}</span>
            </div>
            <div class="meta">${TEXT.statusCode}：${item.status_code || 0} · ${TEXT.duration}：${item.response_time_ms || 0} ms</div>
            <div class="meta">${escapeHTML(item.message || "")}</div>
            <pre class="history-preview">${escapeHTML(item.response_preview || "")}</pre>
        </div>
    `).join("");
}

function renderHealthBars(history) {
    const target = $("healthBars");
    if (!history.length) {
        target.innerHTML = `<div class="feed-item"><div class="meta">${TEXT.healthNoData}</div></div>`;
        return;
    }

    const successCount = history.filter((item) => item.status === "success").length;
    const failedCount = history.filter((item) => item.status === "failed").length;
    const avgDuration = Math.round(
        history.reduce((sum, item) => sum + (item.response_time_ms || 0), 0) / Math.max(history.length, 1)
    );

    const rows = [
        { label: TEXT.healthSuccess, value: successCount, total: history.length, className: "fill-success", suffix: `${successCount}/${history.length}` },
        { label: TEXT.healthFailed, value: failedCount, total: history.length, className: "fill-failed", suffix: `${failedCount}/${history.length}` },
        { label: TEXT.healthAvg, value: Math.min(avgDuration, 1000), total: 1000, className: "fill-neutral", suffix: `${avgDuration} ms` },
    ];

    target.innerHTML = rows.map((row) => {
        const width = row.total > 0 ? Math.max(6, Math.round((row.value / row.total) * 100)) : 0;
        return `
            <div class="health-row">
                <div class="health-label">
                    <span>${row.label}</span>
                    <strong>${row.suffix}</strong>
                </div>
                <div class="health-track">
                    <div class="health-fill ${row.className}" style="width:${width}%"></div>
                </div>
            </div>
        `;
    }).join("");
}

function renderActivityFeed(history) {
    const target = $("activityFeed");
    if (!history.length) {
        target.innerHTML = `<div class="feed-item"><div class="meta">${TEXT.recentFeedEmpty}</div></div>`;
        return;
    }

    target.innerHTML = history.slice(0, 4).map((item) => `
        <div class="feed-item">
            <strong>${escapeHTML(item.task_name)}</strong>
            <div class="meta">${statusLabel(item.status)} · ${escapeHTML(item.created_at)}</div>
            <div class="meta">${escapeHTML(item.message || "")}</div>
        </div>
    `).join("");
}

function fillSettings(data) {
    $("scheduleEnabled").checked = !!data.schedule_config.enabled;
    $("scheduleHour").value = data.schedule_config.hour;
    $("scheduleMinute").value = data.schedule_config.minute;
    $("scheduleWorkers").value = data.schedule_config.max_workers;

    $("telegramEnabled").checked = !!data.notify_config.telegram_enabled;
    $("telegramBotToken").value = data.notify_config.telegram_bot_token || "";
    $("telegramChatID").value = data.notify_config.telegram_chat_id || "";
    $("webhookEnabled").checked = !!data.notify_config.webhook_enabled;
    $("webhookURL").value = data.notify_config.webhook_url || "";
    $("notifyOnSuccess").checked = !!data.notify_config.notify_on_success;
    $("notifyOnFailure").checked = !!data.notify_config.notify_on_failure;

    $("securityEnabled").checked = !!data.security_config.enabled;
}

async function loadBootstrap() {
    const { data } = await api("/api/bootstrap");
    state.tasks = data.tasks;
    state.history = data.history;
    renderSummary(data.stats);
    renderTasks(data.tasks);
    renderHistory(data.history);
    renderHealthBars(data.history);
    renderActivityFeed(data.history);
    fillSettings(data);
}

function getTaskPayload() {
    let headers = {};
    const headersText = $("taskHeaders").value.trim();
    if (headersText) headers = JSON.parse(headersText);
    return {
        name: $("taskName").value.trim(),
        method: $("taskMethod").value,
        url: $("taskURL").value.trim(),
        headers,
        body: $("taskBody").value,
        curl_command: $("taskCurl").value.trim(),
        enabled: $("taskEnabled").checked,
        schedule_enabled: $("taskScheduleEnabled").checked,
        schedule_hour: Number($("taskHour").value),
        schedule_minute: Number($("taskMinute").value),
        timeout_seconds: Number($("taskTimeout").value),
        retry_count: Number($("taskRetry").value),
        success_keywords: $("taskSuccessKeywords").value.trim(),
        failure_keywords: $("taskFailureKeywords").value.trim(),
    };
}

async function parseCurl() {
    const result = await api("/api/tasks/parse", {
        method: "POST",
        body: JSON.stringify(getTaskPayload()),
    });
    $("taskPreview").textContent = JSON.stringify(result.config, null, 2);
    if (!$("taskURL").value.trim()) $("taskURL").value = result.config.url || "";
    if (!$("taskBody").value.trim()) $("taskBody").value = result.config.body || "";
    $("taskHeaders").value = JSON.stringify(result.config.headers || {}, null, 2);
    showMessage(TEXT.parseSuccess);
}

async function saveTask() {
    const payload = getTaskPayload();
    const method = state.editingId ? "PUT" : "POST";
    const url = state.editingId ? `/api/tasks/${state.editingId}` : "/api/tasks";
    const result = await api(url, { method, body: JSON.stringify(payload) });
    $("taskPreview").textContent = JSON.stringify(result.task, null, 2);
    showMessage(state.editingId ? TEXT.taskUpdated : TEXT.taskCreated);
    resetTask(false);
    await loadBootstrap();
}

function editTask(id) {
    const task = state.tasks.find((item) => item.id === id);
    if (!task) return;
    state.editingId = id;
    $("taskFormTitle").textContent = `${TEXT.editTaskPrefix}${id}`;
    $("taskName").value = task.name;
    $("taskMethod").value = task.method;
    $("taskURL").value = task.url;
    $("taskHeaders").value = JSON.stringify(task.headers || {}, null, 2);
    $("taskBody").value = task.body || "";
    $("taskCurl").value = task.curl_command || "";
    $("taskEnabled").checked = !!task.enabled;
    $("taskScheduleEnabled").checked = !!task.schedule_enabled;
    $("taskHour").value = task.schedule_hour;
    $("taskMinute").value = task.schedule_minute;
    $("taskTimeout").value = task.timeout_seconds;
    $("taskRetry").value = task.retry_count;
    $("taskSuccessKeywords").value = task.success_keywords || "";
    $("taskFailureKeywords").value = task.failure_keywords || "";
    $("taskPreview").textContent = JSON.stringify(task, null, 2);
    window.scrollTo({ top: 0, behavior: "smooth" });
}

function resetTask(showToast = true) {
    state.editingId = null;
    $("taskFormTitle").textContent = TEXT.createTask;
    $("taskName").value = "";
    $("taskMethod").value = "GET";
    $("taskURL").value = "";
    $("taskHeaders").value = "";
    $("taskBody").value = "";
    $("taskCurl").value = "";
    $("taskEnabled").checked = true;
    $("taskScheduleEnabled").checked = false;
    $("taskHour").value = 8;
    $("taskMinute").value = 0;
    $("taskTimeout").value = 30;
    $("taskRetry").value = 0;
    $("taskSuccessKeywords").value = "";
    $("taskFailureKeywords").value = "";
    $("taskPreview").textContent = TEXT.taskPreviewHint;
    if (showToast) showMessage(TEXT.formReset);
}

async function runTask(id) {
    await api(`/api/tasks/${id}/run`, { method: "POST" });
    showMessage(TEXT.taskExecuted);
    await loadBootstrap();
}

async function deleteTask(id) {
    if (!window.confirm(TEXT.confirmDelete)) return;
    await api(`/api/tasks/${id}`, { method: "DELETE" });
    showMessage(TEXT.taskDeleted);
    await loadBootstrap();
}

async function runAllTasks() {
    await api("/api/tasks/run-all", { method: "POST" });
    showMessage(TEXT.allExecuted);
    await loadBootstrap();
}

async function clearHistory() {
    if (!window.confirm(TEXT.confirmClear)) return;
    await api("/api/history", { method: "DELETE" });
    showMessage(TEXT.historyCleared);
    await loadBootstrap();
}

async function saveSchedule() {
    await api("/api/settings/schedule", {
        method: "PUT",
        body: JSON.stringify({
            enabled: $("scheduleEnabled").checked,
            hour: Number($("scheduleHour").value),
            minute: Number($("scheduleMinute").value),
            max_workers: Number($("scheduleWorkers").value),
        }),
    });
    showMessage(TEXT.scheduleSaved);
}

async function checkSchedule() {
    await api("/api/settings/schedule/check", { method: "POST" });
    showMessage(TEXT.scheduleChecked);
    await loadBootstrap();
}

async function saveNotify() {
    await api("/api/settings/notify", {
        method: "PUT",
        body: JSON.stringify({
            telegram_enabled: $("telegramEnabled").checked,
            telegram_bot_token: $("telegramBotToken").value.trim(),
            telegram_chat_id: $("telegramChatID").value.trim(),
            webhook_enabled: $("webhookEnabled").checked,
            webhook_url: $("webhookURL").value.trim(),
            notify_on_success: $("notifyOnSuccess").checked,
            notify_on_failure: $("notifyOnFailure").checked,
        }),
    });
    showMessage(TEXT.notifySaved);
}

async function testNotify() {
    await api("/api/settings/notify/test", { method: "POST" });
    showMessage(TEXT.notifyTestSent);
}

async function saveSecurity() {
    await api("/api/settings/security", {
        method: "PUT",
        body: JSON.stringify({
            enabled: $("securityEnabled").checked,
            password: $("securityPassword").value,
        }),
    });
    $("securityPassword").value = "";
    showMessage(TEXT.securitySaved);
}

function bindTaskSearch() {
    $("taskSearch").addEventListener("input", (event) => {
        state.taskSearch = event.target.value;
        renderTasks(state.tasks);
    });

    document.querySelectorAll("#statusFilters .filter-chip").forEach((button) => {
        button.addEventListener("click", () => {
            state.taskFilter = button.dataset.filter;
            document.querySelectorAll("#statusFilters .filter-chip").forEach((item) => {
                item.classList.toggle("filter-chip-active", item === button);
            });
            renderTasks(state.tasks);
        });
    });
}

function attachEvents() {
    $("runAllBtn").addEventListener("click", () => runAllTasks().catch(handleError));
    $("refreshBtn").addEventListener("click", () => loadBootstrap().catch(handleError));
    $("newTaskBtn").addEventListener("click", () => resetTask());
    $("parseCurlBtn").addEventListener("click", () => parseCurl().catch(handleError));
    $("saveTaskBtn").addEventListener("click", () => saveTask().catch(handleError));
    $("resetTaskBtn").addEventListener("click", () => resetTask());
    $("clearHistoryBtn").addEventListener("click", () => clearHistory().catch(handleError));
    $("saveScheduleBtn").addEventListener("click", () => saveSchedule().catch(handleError));
    $("checkScheduleBtn").addEventListener("click", () => checkSchedule().catch(handleError));
    $("saveNotifyBtn").addEventListener("click", () => saveNotify().catch(handleError));
    $("testNotifyBtn").addEventListener("click", () => testNotify().catch(handleError));
    $("saveSecurityBtn").addEventListener("click", () => saveSecurity().catch(handleError));
    bindTaskSearch();
}

function handleError(error) {
    showMessage(error.message || TEXT.unexpectedError, "error");
}

attachEvents();
loadBootstrap().catch(handleError);
