let state = {
    tasks: [],
    history: [],
    scheduleConfig: { enabled: false, hour: 8, minute: 0, second: 0, max_workers: 4 },
    editingId: null,
    taskFilter: "all",
    taskSearch: "",
    currentView: "overview",
    loading: {
        bootstrap: false,
    },
    taskActionKey: "",
};

const TEXT = {
    requestFailed: "请求失败",
    unexpectedError: "发生了未预期的错误",
    totalTasks: "任务总数",
    enabledTasks: "已启用任务",
    recentSuccess: "成功率",
    averageDuration: "平均响应耗时",
    totalTasksNote: "系统中已登记的任务",
    enabledTasksNote: "参与自动调度的任务",
    recentSuccessNote: "近期签到成功比例",
    averageDurationNote: "接口平均响应毫秒数",
    noTasks: "暂无签到任务，点击“新建任务”开始添加。",
    noHistory: "暂无日志记录，任务触发后将在此处实时显示。",
    enabled: "已启用",
    disabled: "已禁用",
    globalSchedule: "跟随全局定时",
    timeout: "超时",
    retry: "重试",
    lastRun: "上次触发",
    nextRun: "下次预定触发",
    neverRun: "未运行",
    noSchedule: "未设置定时",
    lastDuration: "耗时",
    run: "立即运行",
    edit: "编辑",
    del: "删除",
    statusCode: "状态码",
    duration: "耗时",
    parseSuccess: "cURL 解析成功",
    taskCreated: "任务创建成功",
    taskUpdated: "任务更新成功",
    taskDeleted: "任务已删除",
    taskExecuted: "已触发任务执行",
    allExecuted: "已触发所有已启用任务",
    historyCleared: "日志历史记录已清空",
    scheduleSaved: "全局定时设置已保存",
    scheduleChecked: "定时检查已执行",
    notifySaved: "通知设置已保存",
    notifyTestSent: "测试通知已发送",
    securitySaved: "安全设置已保存",
    formReset: "表单已重置",
    createTask: "新建签到任务",
    editTaskPrefix: "编辑任务 #",
    taskPreviewHint: "解析或编辑配置后，生成的请求数据将在此处预览。",
    confirmDelete: "确定要删除此签到任务吗？",
    confirmClear: "确定要清空所有历史日志记录吗？",
    statusSuccess: "成功",
    statusFailed: "失败",
    statusIdle: "未运行",
    healthSuccess: "签到成功次数",
    healthFailed: "签到失败次数",
    healthAvg: "平均响应耗时",
    healthNoData: "暂无统计数据",
    recentFeedEmpty: "近期无活动记录",
    loadingData: "加载中...",
    loadingAction: "处理中...",
    taskSchedulePrefix: "独立 ",
    aggressiveMode: "抢零点模式",
    requestStartedAt: "发起时间",
};

const $ = (id) => document.getElementById(id);
const messageEl = $("message");

function setView(view) {
    state.currentView = view;

    const views = {
        overview: $("overviewView"),
        tasks: $("tasksView"),
        taskEditor: $("taskEditorView"),
        settings: $("settingsView"),
    };

    const buttons = {
        overview: $("navOverviewBtn"),
        tasks: $("navTasksBtn"),
        settings: $("navSettingsBtn"),
    };

    Object.entries(views).forEach(([key, element]) => {
        if (element) element.classList.toggle("hidden", key !== view);
    });

    const navView = view === "taskEditor" ? "tasks" : view;

    Object.entries(buttons).forEach(([key, element]) => {
        if (element) element.classList.toggle("sidebar-btn-active", key === navView);
    });

    const pageTitles = { overview: "总览控制台", tasks: "任务中心", taskEditor: "任务配置", settings: "全局设置" };
    const titleEl = $("pageTitle");
    if (titleEl) titleEl.textContent = pageTitles[view] || "总览控制台";

    renderTopbarActions();
}

function renderTopbarActions() {
    const view = state.currentView;
    const container = $("topbarActions");
    if (!container) return;

    let html = "";
    if (view === "overview") {
        html = `<button class="btn-ghost" id="refreshBtn">刷新数据</button>
                <button class="btn-ghost" id="newTaskBtn">+ 新建任务</button>
                <button class="btn-primary" id="runAllBtn">运行所有已启用任务</button>`;
    } else if (view === "tasks") {
        html = `<button class="btn-primary" id="taskPageNewBtn">+ 新建任务</button>
                <button class="btn-ghost" id="runAllBtn">运行所有已启用任务</button>`;
    } else if (view === "taskEditor") {
        html = `<button class="btn-ghost" id="taskEditorBackBtn">&larr; 返回任务列表</button>`;
    }
    container.innerHTML = html;
    rebindTopbarEvents();
}

function rebindTopbarEvents() {
    const refreshBtn = $("refreshBtn");
    if (refreshBtn) refreshBtn.addEventListener("click", (event) => withButtonLoading(event.currentTarget, () => loadBootstrap().catch(handleError), TEXT.loadingData));

    const newTaskBtn = $("newTaskBtn");
    if (newTaskBtn) newTaskBtn.addEventListener("click", () => { resetTask(false); openTaskEditor(); });

    const runAllBtn = $("runAllBtn");
    if (runAllBtn) runAllBtn.addEventListener("click", (event) => withButtonLoading(event.currentTarget, () => runAllTasks().catch(handleError)));

    const taskPageNewBtn = $("taskPageNewBtn");
    if (taskPageNewBtn) taskPageNewBtn.addEventListener("click", () => { resetTask(false); openTaskEditor(); });

    const taskEditorBackBtn = $("taskEditorBackBtn");
    if (taskEditorBackBtn) taskEditorBackBtn.addEventListener("click", () => setView("tasks"));
}

function openTaskEditor() {
    setView("taskEditor");
    window.scrollTo({ top: 0, behavior: "smooth" });
}

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
    if (!messageEl) return;
    messageEl.innerHTML = `
        <div class="toast-content">
            <span class="toast-icon">${type === "success" ? "✓" : "!"}</span>
            <span>${escapeHTML(text)}</span>
        </div>
        <div class="toast-progress"></div>
    `;
    messageEl.className = `toast toast-${type}`;
    window.clearTimeout(showMessage.timer);
    showMessage.timer = window.setTimeout(() => {
        messageEl.className = "toast hidden";
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

function extractHost(urlStr) {
    if (!urlStr) return "endpoint";
    try {
        const u = new URL(urlStr);
        return u.hostname;
    } catch (e) {
        return urlStr.length > 25 ? urlStr.substring(0, 25) + "..." : urlStr;
    }
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

function setButtonLoading(button, isLoading, loadingText = TEXT.loadingAction) {
    if (!button) return;
    if (isLoading) {
        if (!button.dataset.originalText) {
            button.dataset.originalText = button.textContent;
        }
        button.disabled = true;
        button.classList.add("is-loading");
        button.textContent = loadingText;
        return;
    }

    button.disabled = false;
    button.classList.remove("is-loading");
    if (button.dataset.originalText) {
        button.textContent = button.dataset.originalText;
        delete button.dataset.originalText;
    }
}

async function withButtonLoading(button, action, loadingText) {
    setButtonLoading(button, true, loadingText);
    try {
        return await action();
    } finally {
        setButtonLoading(button, false, loadingText);
    }
}

function formatMethod(method) {
    const m = (method || "GET").toUpperCase();
    const cls = `method-${m.toLowerCase()}`;
    return `<span class="method-badge ${cls}">${escapeHTML(m)}</span>`;
}

function getTaskActionMarkup(task) {
    const runBusy = state.taskActionKey === `run:${task.id}`;
    const deleteBusy = state.taskActionKey === `delete:${task.id}`;

    return `
        <button class="btn-primary ${runBusy ? "is-loading" : ""}" data-task-action="run" data-task-id="${task.id}" ${runBusy ? "disabled" : ""}>${runBusy ? TEXT.loadingAction : TEXT.run}</button>
        <button class="btn-secondary" data-task-action="edit" data-task-id="${task.id}">${TEXT.edit}</button>
        <button class="btn-danger ${deleteBusy ? "is-loading" : ""}" data-task-action="delete" data-task-id="${task.id}" ${deleteBusy ? "disabled" : ""}>${deleteBusy ? TEXT.loadingAction : TEXT.del}</button>
    `;
}

function formatDateTime(value) {
    return [
        value.getFullYear(),
        pad(value.getMonth() + 1),
        pad(value.getDate()),
    ].join("-") + " " + [
        pad(value.getHours()),
        pad(value.getMinutes()),
        pad(value.getSeconds()),
    ].join(":");
}

function getTaskSchedule(task) {
    if (task.schedule_enabled) {
        return {
            enabled: true,
            hour: Number(task.schedule_hour || 0),
            minute: Number(task.schedule_minute || 0),
            second: Number(task.schedule_second || 0),
        };
    }

    if (state.scheduleConfig.enabled) {
        return {
            enabled: true,
            hour: Number(state.scheduleConfig.hour || 0),
            minute: Number(state.scheduleConfig.minute || 0),
            second: Number(state.scheduleConfig.second || 0),
        };
    }

    return { enabled: false, hour: 0, minute: 0, second: 0 };
}

function getNextRunLabel(task) {
    if (!task.enabled) return TEXT.disabled;

    const schedule = getTaskSchedule(task);
    if (!schedule.enabled) return TEXT.noSchedule;

    const now = new Date();
    const nextRun = new Date(now);
    nextRun.setHours(schedule.hour, schedule.minute, schedule.second, 0);
    if (nextRun <= now) {
        nextRun.setDate(nextRun.getDate() + 1);
    }
    return formatDateTime(nextRun);
}

function renderSummary(stats) {
    const grid = $("summaryGrid");
    if (!grid) return;

    const historyCount = state.history ? state.history.length : 0;
    const successRate = historyCount > 0
        ? Math.round(((stats.recent_success || 0) / historyCount) * 100)
        : 100;

    const cards = [
        [TEXT.totalTasks, stats.total_tasks, TEXT.totalTasksNote],
        [TEXT.enabledTasks, stats.enabled_tasks, TEXT.enabledTasksNote],
        [TEXT.recentSuccess, `${successRate}%`, TEXT.recentSuccessNote],
        [TEXT.averageDuration, `${stats.avg_duration_ms} ms`, TEXT.averageDurationNote],
    ];

    grid.innerHTML = cards.map(([label, value, note]) => `
        <div class="stat-card">
            <span class="stat-label">${label}</span>
            <div class="stat-value">${value}</div>
            <span class="stat-sub">${note}</span>
        </div>
    `).join("");
}

function getFilteredTasks(tasks) {
    return tasks.filter((task) => {
        const search = state.taskSearch.trim().toLowerCase();
        const name = String(task.name || "").toLowerCase();
        const url = String(task.url || "").toLowerCase();
        const matchesSearch = !search || name.includes(search) || url.includes(search);

        let matchesFilter = true;
        if (state.taskFilter === "active") {
            matchesFilter = !!task.enabled;
        } else if (state.taskFilter !== "all") {
            matchesFilter = (task.last_status || "idle") === state.taskFilter;
        }

        return matchesSearch && matchesFilter;
    });
}

function renderTasks(tasks) {
    const filtered = getFilteredTasks(tasks);
    const target = $("taskList");
    if (!target) return;

    if (!filtered.length) {
        target.innerHTML = `<div class="card" style="grid-column:1/-1;text-align:center;padding:32px"><div style="font-weight:600;margin-bottom:4px">没有匹配的任务</div><div style="color:var(--text-muted);font-size:13px">${TEXT.noTasks}</div></div>`;
        return;
    }

    target.innerHTML = filtered.map((task) => `
        <div class="task-card">
            <div class="task-card-head">
                <div>
                    <div style="display:flex;align-items:center;gap:8px">
                        <div class="task-title">${escapeHTML(task.name)}</div>
                        ${formatMethod(task.method)}
                    </div>
                    <div class="task-url" title="${escapeHTML(task.url)}">${escapeHTML(extractHost(task.url))} &bull; ${escapeHTML(task.url)}</div>
                </div>
                <label class="toggle" style="margin:0" title="切换启用状态">
                    <input type="checkbox" ${task.enabled ? "checked" : ""} data-toggle-task-id="${task.id}">
                </label>
            </div>

            <div class="task-card-meta">
                <span class="section-tag" style="margin:0;padding:2px 8px;border-radius:var(--radius-sm);background:rgba(255,255,255,0.05)">${statusLabel(task.last_status)}</span>
                <span>定时: ${task.schedule_enabled ? `${pad(task.schedule_hour)}:${pad(task.schedule_minute)}:${pad(task.schedule_second)}` : TEXT.globalSchedule}</span>
                <span>${TEXT.timeout}: ${task.timeout_seconds}s</span>
                <span>${TEXT.retry}: ${task.retry_count}</span>
                ${task.aggressive_mode ? `<span>${TEXT.aggressiveMode}</span>` : ""}
            </div>

            <div style="font-size:12px;color:var(--text-muted);display:flex;flex-direction:column;gap:2px">
                <div>${TEXT.lastRun}: ${task.last_run_at || TEXT.neverRun}</div>
                <div>${TEXT.nextRun}: ${getNextRunLabel(task)}</div>
                <div>${TEXT.lastDuration}: ${task.last_duration_ms || 0} ms</div>
            </div>

            <div class="task-card-actions">
                ${getTaskActionMarkup(task)}
            </div>
        </div>
    `).join("");
}

function renderHistory(history) {
    const target = $("historyList");
    if (!target) return;

    if (!history.length) {
        target.innerHTML = `<div class="card" style="text-align:center;padding:32px"><div style="font-weight:600;margin-bottom:4px">暂无运行日志</div><div style="color:var(--text-muted);font-size:13px">${TEXT.noHistory}</div></div>`;
        return;
    }

    target.innerHTML = history.map((item) => {
        const isSuccess = item.status === "success";
        return `
            <div class="history-item">
                <div class="history-main-info">
                    <span class="history-badge ${isSuccess ? "status-2xx" : "status-err"}">${item.status_code || (isSuccess ? 200 : 500)}</span>
                    <div>
                        <strong style="font-size:14px;color:var(--text);display:block">${escapeHTML(item.task_name)}</strong>
                        <span style="font-size:12px;color:var(--text-muted)">${escapeHTML(item.triggered_by)} &bull; ${escapeHTML(item.created_at)}</span>
                    </div>
                </div>
                <div style="text-align:right">
                    <span style="font-size:13px;font-family:monospace;color:var(--text-soft)">${item.response_time_ms || 0} ms</span>
                    <div style="font-size:12px;color:var(--text-muted)">${escapeHTML(item.message || "")}</div>
                </div>
            </div>
        `;
    }).join("");
}

function renderHealthBars(history) {
    const target = $("healthBars");
    if (!target) return;

    if (!history.length) {
        target.innerHTML = `<div style="color:var(--text-muted);font-size:13px">${TEXT.healthNoData}</div>`;
        return;
    }

    const successCount = history.filter((item) => item.status === "success").length;
    const failedCount = history.filter((item) => item.status === "failed").length;
    const avgDuration = Math.round(
        history.reduce((sum, item) => sum + (item.response_time_ms || 0), 0) / Math.max(history.length, 1)
    );

    const rows = [
        { label: TEXT.healthSuccess, value: successCount, total: history.length, isSuccess: true, suffix: `${successCount}/${history.length}` },
        { label: TEXT.healthFailed, value: failedCount, total: history.length, isSuccess: false, suffix: `${failedCount}/${history.length}` },
        { label: TEXT.healthAvg, value: Math.min(avgDuration, 1000), total: 1000, isSuccess: true, suffix: `${avgDuration} ms` },
    ];

    target.innerHTML = rows.map((row) => {
        const width = row.total > 0 ? Math.max(6, Math.round((row.value / row.total) * 100)) : 0;
        const barClass = row.isSuccess ? "health-bar-fill-success" : "health-bar-fill-failed";
        return `
            <div class="health-bar-item">
                <div class="health-bar-label">${row.label}</div>
                <div class="health-bar-track">
                    <div class="${barClass}" style="width:${width}%"></div>
                </div>
                <div class="health-bar-meta">${row.suffix}</div>
            </div>
        `;
    }).join("");
}

function renderActivityFeed(history) {
    const target = $("activityFeed");
    if (!target) return;

    if (!history.length) {
        target.innerHTML = `<div style="color:var(--text-muted);font-size:13px">${TEXT.recentFeedEmpty}</div>`;
        return;
    }

    target.innerHTML = history.slice(0, 4).map((item) => `
        <div class="feed-item">
            <div class="feed-left">
                <span class="section-tag" style="margin:0">${statusLabel(item.status)}</span>
                <strong>${escapeHTML(item.task_name)}</strong>
            </div>
            <span class="feed-time">${escapeHTML(item.created_at)}</span>
        </div>
    `).join("");
}

function fillSettings(data) {
    state.scheduleConfig = data.schedule_config || state.scheduleConfig;
    $("scheduleEnabled").checked = !!data.schedule_config.enabled;
    $("scheduleHour").value = data.schedule_config.hour;
    $("scheduleMinute").value = data.schedule_config.minute;
    $("scheduleSecond").value = data.schedule_config.second;
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
    state.loading.bootstrap = true;
    const { data } = await api("/api/bootstrap");
    state.tasks = data.tasks;
    state.history = data.history;
    state.scheduleConfig = data.schedule_config || state.scheduleConfig;
    renderSummary(data.stats);
    renderTasks(data.tasks);
    renderHistory(data.history);
    renderHealthBars(data.history);
    renderActivityFeed(data.history);
    fillSettings(data);
    state.loading.bootstrap = false;
}

function formatJSONText(text) {
    if (!text || !text.trim()) return text;
    try {
        const obj = JSON.parse(text);
        return JSON.stringify(obj, null, 2);
    } catch (e) {
        return text;
    }
}

function getTaskPayload() {
    let headers = {};
    const headersText = $("taskHeaders").value.trim();
    if (headersText) {
        try {
            headers = JSON.parse(headersText);
        } catch (e) {
            throw new Error("请求头 JSON 格式不正确");
        }
    }

    return {
        name: $("taskName").value.trim(),
        method: $("taskMethod").value,
        url: $("taskURL").value.trim(),
        headers,
        body: $("taskBody").value,
        curl_command: $("taskCurl").value.trim(),
        enabled: $("taskEnabled").checked,
        schedule_enabled: $("taskScheduleEnabled").checked,
        aggressive_mode: $("taskAggressiveMode").checked,
        schedule_hour: Number($("taskHour").value),
        schedule_minute: Number($("taskMinute").value),
        schedule_second: Number($("taskSecond").value),
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
    setView("tasks");
}

function editTask(id) {
    const task = state.tasks.find((item) => item.id === id);
    if (!task) return;

    openTaskEditor();
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
    $("taskAggressiveMode").checked = !!task.aggressive_mode;
    $("taskHour").value = task.schedule_hour;
    $("taskMinute").value = task.schedule_minute;
    $("taskSecond").value = task.schedule_second ?? 0;
    $("taskTimeout").value = task.timeout_seconds;
    $("taskRetry").value = task.retry_count;
    $("taskSuccessKeywords").value = task.success_keywords || "";
    $("taskFailureKeywords").value = task.failure_keywords || "";
    $("taskPreview").textContent = JSON.stringify(task, null, 2);
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
    $("taskAggressiveMode").checked = false;
    $("taskHour").value = 8;
    $("taskMinute").value = 0;
    $("taskSecond").value = 0;
    $("taskTimeout").value = 30;
    $("taskRetry").value = 0;
    $("taskSuccessKeywords").value = "";
    $("taskFailureKeywords").value = "";
    $("taskPreview").textContent = TEXT.taskPreviewHint;

    if (showToast) showMessage(TEXT.formReset);
}

async function toggleTaskActive(id, enabled) {
    const task = state.tasks.find((t) => t.id === id);
    if (!task) return;
    const payload = { ...task, enabled };
    await api(`/api/tasks/${id}`, { method: "PUT", body: JSON.stringify(payload) });
    task.enabled = enabled;
    showMessage(enabled ? "任务已启用" : "任务已禁用");
    await loadBootstrap();
}

async function runTask(id) {
    state.taskActionKey = `run:${id}`;
    renderTasks(state.tasks);
    try {
        await api(`/api/tasks/${id}/run`, { method: "POST" });
        showMessage(TEXT.taskExecuted);
        await loadBootstrap();
    } finally {
        state.taskActionKey = "";
        renderTasks(state.tasks);
    }
}

async function deleteTask(id) {
    if (!window.confirm(TEXT.confirmDelete)) return;
    state.taskActionKey = `delete:${id}`;
    renderTasks(state.tasks);
    try {
        await api(`/api/tasks/${id}`, { method: "DELETE" });
        showMessage(TEXT.taskDeleted);
        await loadBootstrap();
    } finally {
        state.taskActionKey = "";
        renderTasks(state.tasks);
    }
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
            second: Number($("scheduleSecond").value),
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
    const searchEl = $("taskSearch");
    if (searchEl) {
        searchEl.addEventListener("input", (event) => {
            state.taskSearch = event.target.value;
            renderTasks(state.tasks);
        });
        document.addEventListener("keydown", (e) => {
            if (e.key === "/" && document.activeElement !== searchEl && document.activeElement.tagName !== "INPUT" && document.activeElement.tagName !== "TEXTAREA") {
                e.preventDefault();
                searchEl.focus();
            }
        });
    }

    document.querySelectorAll("#statusFilters .filter-chip").forEach((button) => {
        button.addEventListener("click", () => {
            state.taskFilter = button.dataset.filter;
            document.querySelectorAll("#statusFilters .filter-chip").forEach((item) => {
                item.classList.toggle("filter-chip-active", item === button);
            });
            renderTasks(state.tasks);
        });
    });

    const taskListEl = $("taskList");
    if (taskListEl) {
        taskListEl.addEventListener("change", (event) => {
            const toggleInput = event.target.closest("input[data-toggle-task-id]");
            if (toggleInput) {
                const id = Number(toggleInput.dataset.toggleTaskId);
                toggleTaskActive(id, toggleInput.checked).catch(handleError);
            }
        });

        taskListEl.addEventListener("click", (event) => {
            const button = event.target.closest("button[data-task-action]");
            if (!button) return;

            const id = Number(button.dataset.taskId);
            const action = button.dataset.taskAction;

            if (action === "run") {
                runTask(id).catch(handleError);
                return;
            }

            if (action === "edit") {
                editTask(id);
                return;
            }

            if (action === "delete") {
                deleteTask(id).catch(handleError);
            }
        });
    }
}

function attachEvents() {
    const navOverview = $("navOverviewBtn");
    if (navOverview) navOverview.addEventListener("click", () => setView("overview"));
    const navTasks = $("navTasksBtn");
    if (navTasks) navTasks.addEventListener("click", () => setView("tasks"));
    const navSettings = $("navSettingsBtn");
    if (navSettings) navSettings.addEventListener("click", () => setView("settings"));

    const saveTaskEl = $("saveTaskBtn");
    if (saveTaskEl) saveTaskEl.addEventListener("click", (event) => withButtonLoading(event.currentTarget, () => saveTask().catch(handleError)));

    const taskEditorNewBlank = $("taskEditorNewBlankBtn");
    if (taskEditorNewBlank) taskEditorNewBlank.addEventListener("click", () => resetTask());

    const parseCurlEl = $("parseCurlBtn");
    if (parseCurlEl) parseCurlEl.addEventListener("click", (event) => withButtonLoading(event.currentTarget, () => parseCurl().catch(handleError)));

    const resetTaskEl = $("resetTaskBtn");
    if (resetTaskEl) resetTaskEl.addEventListener("click", () => resetTask());

    const clearHistoryEl = $("clearHistoryBtn");
    if (clearHistoryEl) clearHistoryEl.addEventListener("click", (event) => withButtonLoading(event.currentTarget, () => clearHistory().catch(handleError)));

    const saveScheduleEl = $("saveScheduleBtn");
    if (saveScheduleEl) saveScheduleEl.addEventListener("click", (event) => withButtonLoading(event.currentTarget, () => saveSchedule().catch(handleError)));

    const checkScheduleEl = $("checkScheduleBtn");
    if (checkScheduleEl) checkScheduleEl.addEventListener("click", (event) => withButtonLoading(event.currentTarget, () => checkSchedule().catch(handleError)));

    const saveNotifyEl = $("saveNotifyBtn");
    if (saveNotifyEl) saveNotifyEl.addEventListener("click", (event) => withButtonLoading(event.currentTarget, () => saveNotify().catch(handleError)));

    const testNotifyEl = $("testNotifyBtn");
    if (testNotifyEl) testNotifyEl.addEventListener("click", (event) => withButtonLoading(event.currentTarget, () => testNotify().catch(handleError)));

    const saveSecurityEl = $("saveSecurityBtn");
    if (saveSecurityEl) saveSecurityEl.addEventListener("click", (event) => withButtonLoading(event.currentTarget, () => saveSecurity().catch(handleError)));

    const taskHeadersEl = $("taskHeaders");
    if (taskHeadersEl) {
        taskHeadersEl.addEventListener("blur", () => {
            taskHeadersEl.value = formatJSONText(taskHeadersEl.value);
        });
    }

    bindTaskSearch();

    const mobileBtn = $("mobileMenuBtn");
    const sidebar = $("sidebar");
    if (mobileBtn && sidebar) {
        mobileBtn.addEventListener("click", () => {
            sidebar.classList.toggle("open");
        });
        const mainContent = document.querySelector(".main-content");
        if (mainContent) {
            mainContent.addEventListener("click", () => {
                sidebar.classList.remove("open");
            });
        }
    }
}

function handleError(error) {
    showMessage(error.message || TEXT.unexpectedError, "error");
}

attachEvents();
setView("overview");
renderTopbarActions();
loadBootstrap().catch(handleError);
