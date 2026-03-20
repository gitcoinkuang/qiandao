let state = {
    tasks: [],
    history: [],
    editingId: null,
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
        throw new Error(data.error || "Request failed");
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

function statusClass(status) {
    if (status === "success") return "status-success";
    if (status === "failed") return "status-failed";
    return "status-idle";
}

function renderSummary(stats) {
    $("summaryGrid").innerHTML = [
        ["Tasks", stats.total_tasks],
        ["Enabled", stats.enabled_tasks],
        ["Recent success", stats.recent_success],
        ["Avg duration", `${stats.avg_duration_ms} ms`],
    ].map(([label, value]) => `
        <div class="summary-card">
            <span>${label}</span>
            <strong>${value}</strong>
        </div>
    `).join("");
}

function renderTasks(tasks) {
    const target = $("taskList");
    if (!tasks.length) {
        target.innerHTML = `<div class="task-card"><div class="meta">No tasks yet. Create your first one from the form on the right.</div></div>`;
        return;
    }
    target.innerHTML = tasks.map((task) => `
        <div class="task-card">
            <div class="task-top">
                <div>
                    <strong>${escapeHTML(task.name)}</strong>
                    <div class="meta">${escapeHTML(task.method)} · ${escapeHTML(task.url)}</div>
                </div>
                <span class="status-badge ${statusClass(task.last_status)}">${escapeHTML(task.last_status || "idle")}</span>
            </div>
            <div class="pill-row">
                <span class="pill">${task.enabled ? "enabled" : "disabled"}</span>
                <span class="pill">${task.schedule_enabled ? `task schedule ${String(task.schedule_hour).padStart(2, "0")}:${String(task.schedule_minute).padStart(2, "0")}` : "global schedule"}</span>
                <span class="pill">timeout ${task.timeout_seconds}s</span>
                <span class="pill">retry ${task.retry_count}</span>
            </div>
            <div class="meta">Last run: ${task.last_run_at || "never"} · Last duration: ${task.last_duration_ms || 0} ms</div>
            <div class="task-actions">
                <button class="primary" onclick="runTask(${task.id})">Run</button>
                <button class="secondary" onclick="editTask(${task.id})">Edit</button>
                <button class="danger" onclick="deleteTask(${task.id})">Delete</button>
            </div>
        </div>
    `).join("");
}

function renderHistory(history) {
    const target = $("historyList");
    if (!history.length) {
        target.innerHTML = `<div class="history-card"><div class="meta">No history yet.</div></div>`;
        return;
    }
    target.innerHTML = history.map((item) => `
        <div class="history-card">
            <strong>${escapeHTML(item.task_name)}</strong>
            <div class="meta">${escapeHTML(item.status)} · ${escapeHTML(item.triggered_by)} · ${escapeHTML(item.created_at)}</div>
            <div class="meta">Status code: ${item.status_code || 0} · Duration: ${item.response_time_ms || 0} ms</div>
            <div class="meta">${escapeHTML(item.message || "")}</div>
            <pre class="history-preview">${escapeHTML(item.response_preview || "")}</pre>
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
}

async function saveTask() {
    const payload = getTaskPayload();
    const method = state.editingId ? "PUT" : "POST";
    const url = state.editingId ? `/api/tasks/${state.editingId}` : "/api/tasks";
    const result = await api(url, { method, body: JSON.stringify(payload) });
    $("taskPreview").textContent = JSON.stringify(result.task, null, 2);
    showMessage(state.editingId ? "Task updated." : "Task created.");
    resetTask(false);
    await loadBootstrap();
}

function editTask(id) {
    const task = state.tasks.find((item) => item.id === id);
    if (!task) return;
    state.editingId = id;
    $("taskFormTitle").textContent = `Edit task #${id}`;
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
    $("taskFormTitle").textContent = "Create task";
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
    $("taskPreview").textContent = "Parsed task details will appear here.";
    if (showToast) showMessage("Task form reset.");
}

async function runTask(id) {
    await api(`/api/tasks/${id}/run`, { method: "POST" });
    showMessage("Task executed.");
    await loadBootstrap();
}

async function deleteTask(id) {
    if (!window.confirm("Delete this task?")) return;
    await api(`/api/tasks/${id}`, { method: "DELETE" });
    showMessage("Task deleted.");
    await loadBootstrap();
}

async function runAllTasks() {
    await api("/api/tasks/run-all", { method: "POST" });
    showMessage("All enabled tasks executed.");
    await loadBootstrap();
}

async function clearHistory() {
    if (!window.confirm("Clear all history?")) return;
    await api("/api/history", { method: "DELETE" });
    showMessage("History cleared.");
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
    showMessage("Schedule settings saved.");
}

async function checkSchedule() {
    await api("/api/settings/schedule/check", { method: "POST" });
    showMessage("Schedule check completed.");
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
    showMessage("Notification settings saved.");
}

async function testNotify() {
    await api("/api/settings/notify/test", { method: "POST" });
    showMessage("Test notification sent.");
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
    showMessage("Security settings saved.");
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
}

function handleError(error) {
    showMessage(error.message || "Unexpected error", "error");
}

attachEvents();
loadBootstrap().catch(handleError);
