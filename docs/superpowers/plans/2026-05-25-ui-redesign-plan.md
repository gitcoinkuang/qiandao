# QianDao V2 UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 全面重写 QianDao V2 的前端样式和模板，将浅蓝风格改为 Slate Minimal 深色侧边栏卡片风格

**Architecture:** 保持原生 CSS + Vanilla JS + Go html/template 技术栈，不引入任何外部依赖。CSS 使用 CSS 自定义属性作为设计 tokens，HTML 模板重构为侧边栏布局，JS 更新导航逻辑。

**Tech Stack:** CSS Custom Properties, Vanilla JS, Go html/template, Google Fonts (Fira Sans + Fira Code)

**文件变更总览:**
- Rewrite: `static/app.css` (完整重写 ~1400行)
- Rewrite: `templates/index.html` (重构布局 ~450行)
- Modify: `templates/login.html` (更新样式 ~55行)
- Modify: `static/app.js` (更新导航逻辑 ~20行改动)

---

### Task 1: CSS - 设计系统与基础样式（完整重写 app.css）

**Files:**
- Rewrite: `static/app.css` (第1-50行：设计 tokens + 基础 reset)

- [ ] **Step 1: 编写 CSS 自定义属性（Design Tokens）**

替换 `:root` 中的所有变量：

```css
:root {
    --bg-deepest: #080C14;
    --bg: #0B0F1A;
    --bg-elevated: #111627;
    --bg-card: #1A1F2E;
    --bg-card-hover: #212738;
    --line: rgba(255, 255, 255, 0.06);
    --line-strong: rgba(255, 255, 255, 0.10);
    --text: #F1F5F9;
    --text-soft: #94A3B8;
    --text-muted: #64748B;
    --text-inverse: #0B0F1A;
    --primary: #22D3EE;
    --primary-soft: rgba(34, 211, 238, 0.10);
    --primary-strong: #06B6D4;
    --success: #22C55E;
    --success-soft: rgba(34, 197, 94, 0.12);
    --danger: #EF4444;
    --danger-soft: rgba(239, 68, 68, 0.12);
    --amber: #F59E0B;
    --amber-soft: rgba(245, 158, 11, 0.12);
    --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.24);
    --shadow-md: 0 4px 12px rgba(0, 0, 0, 0.28);
    --shadow-lg: 0 12px 40px rgba(0, 0, 0, 0.36);
    --radius-sm: 6px;
    --radius-md: 8px;
    --radius-lg: 12px;
    --radius-xl: 16px;
    --sidebar-w: 240px;
    --sidebar-collapsed-w: 72px;
}
```

- [ ] **Step 2: 编写基础 reset 与 body 样式**

```css
* { box-sizing: border-box; }
html { scroll-behavior: smooth; }
body {
    margin: 0;
    min-height: 100vh;
    background: var(--bg);
    color: var(--text);
    font-family: "Fira Sans", "Noto Sans SC", "Microsoft YaHei", sans-serif;
    overflow-x: hidden;
}
code, pre, .mono, .brand-mark, .stat-value {
    font-family: "Fira Code", "JetBrains Mono", monospace;
}
h1, h2, h3, p { margin: 0; }
a { color: var(--primary); text-decoration: none; }
a:hover { text-decoration: underline; }
```

- [ ] **Step 3: 编写布局容器与侧边栏样式**

```css
.shell {
    display: flex;
    min-height: 100vh;
}
.sidebar {
    position: fixed;
    top: 0;
    left: 0;
    width: var(--sidebar-w);
    height: 100vh;
    background: var(--bg-elevated);
    border-right: 1px solid var(--line-strong);
    display: flex;
    flex-direction: column;
    z-index: 100;
    transition: width 0.2s ease, transform 0.2s ease;
}
.sidebar-brand {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 20px 20px 16px;
    border-bottom: 1px solid var(--line);
}
.brand-mark {
    width: 42px; height: 42px;
    display: grid;
    place-items: center;
    border-radius: var(--radius-md);
    background: linear-gradient(135deg, #22D3EE, #06B6D4);
    color: #0B0F1A;
    font-weight: 700;
    font-size: 18px;
    flex-shrink: 0;
}
.brand-name { font-size: 16px; font-weight: 700; color: var(--text); }
.brand-sub { font-size: 12px; color: var(--text-muted); }
.sidebar-nav { flex: 1; padding: 12px; display: flex; flex-direction: column; gap: 2px; }
.sidebar-btn {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 14px;
    border-radius: var(--radius-md);
    border: none;
    background: transparent;
    color: var(--text-soft);
    font: inherit;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.15s ease, color 0.15s ease;
    width: 100%;
    text-align: left;
    position: relative;
}
.sidebar-btn:hover { background: rgba(255,255,255,0.04); color: var(--text); }
.sidebar-btn-active {
    color: var(--primary);
    background: var(--primary-soft);
}
.sidebar-btn-active::before {
    content: "";
    position: absolute;
    left: -12px;
    top: 50%; transform: translateY(-50%);
    width: 3px;
    height: 20px;
    border-radius: 0 3px 3px 0;
    background: var(--primary);
}
.sidebar-footer {
    padding: 16px 20px;
    border-top: 1px solid var(--line);
    font-size: 12px;
    color: var(--text-muted);
}
.sidebar-footer .status-dot {
    display: inline-block;
    width: 8px; height: 8px;
    border-radius: 50%;
    margin-right: 6px;
    background: var(--success);
}
.main-content {
    margin-left: var(--sidebar-w);
    flex: 1;
    min-height: 100vh;
    padding: 24px 32px;
    min-width: 0;
}
.topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--line);
}
.topbar-left { display: flex; align-items: center; gap: 10px; }
.topbar-title { font-size: 20px; font-weight: 700; }
.topbar-actions { display: flex; gap: 8px; align-items: center; }
```

- [ ] **Step 4: 编写按钮组件样式**

```css
button {
    appearance: none;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    padding: 10px 16px;
    font: inherit;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
    display: inline-flex;
    align-items: center;
    gap: 6px;
}
.btn-primary {
    background: var(--primary);
    color: #0B0F1A;
    border-color: var(--primary);
}
.btn-primary:hover { background: var(--primary-strong); }
.btn-secondary {
    background: transparent;
    color: var(--primary);
    border-color: var(--primary-soft);
}
.btn-secondary:hover { background: var(--primary-soft); }
.btn-ghost {
    background: transparent;
    color: var(--text-soft);
    border-color: var(--line-strong);
}
.btn-ghost:hover { background: rgba(255,255,255,0.04); color: var(--text); }
.btn-danger {
    background: transparent;
    color: var(--danger);
    border-color: var(--danger-soft);
}
.btn-danger:hover { background: var(--danger-soft); }
button:disabled { opacity: 0.5; cursor: not-allowed; }
```

- [ ] **Step 5: 编写卡片组件样式**

```css
.card {
    background: var(--bg-card);
    border: 1px solid var(--line-strong);
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-sm);
}
.card-hover:hover {
    background: var(--bg-card-hover);
    border-color: rgba(255,255,255,0.14);
    box-shadow: var(--shadow-md);
}
```

- [ ] **Step 6: 编写表单控件样式**

```css
input, select, textarea {
    width: 100%;
    padding: 12px 14px;
    border-radius: var(--radius-md);
    border: 1px solid var(--line-strong);
    background: #1E293B;
    color: var(--text);
    font: inherit;
    font-size: 14px;
    transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
input:focus, select:focus, textarea:focus {
    outline: none;
    border-color: var(--primary);
    box-shadow: 0 0 0 3px var(--primary-soft);
}
input::placeholder, textarea::placeholder { color: var(--text-muted); }
.field { display: flex; flex-direction: column; gap: 6px; }
.field > span { font-size: 13px; font-weight: 600; color: var(--text-soft); }
```

- [ ] **Step 7: 编写开关/复选框自定义样式**

```css
.toggle { display: inline-flex; align-items: center; gap: 8px; cursor: pointer; font-size: 14px; color: var(--text-soft); }
.toggle input[type="checkbox"] {
    appearance: none;
    width: 18px; height: 18px;
    border: 1px solid var(--line-strong);
    border-radius: 4px;
    background: #1E293B;
    cursor: pointer;
    transition: all 0.15s ease;
    display: grid;
    place-items: center;
    flex-shrink: 0;
}
.toggle input[type="checkbox"]:checked {
    background: var(--primary);
    border-color: var(--primary);
}
.toggle input[type="checkbox"]:checked::after {
    content: "";
    width: 5px; height: 9px;
    border: solid #0B0F1A;
    border-width: 0 2px 2px 0;
    transform: rotate(45deg);
    margin-top: -1px;
}
```

- [ ] **Step 8: 编写 Toast、Badge、Pill、Skeleton 等组件样式**

```css
/* Toast */
.toast {
    position: fixed; top: 20px; right: 20px;
    z-index: 999;
    min-width: 280px; max-width: 420px;
    padding: 14px 18px;
    border-radius: var(--radius-lg);
    background: var(--bg-card);
    border: 1px solid var(--line-strong);
    box-shadow: var(--shadow-lg);
    font-weight: 600;
    font-size: 14px;
    animation: toast-in 0.2s ease;
}
.toast-success { border-left: 4px solid var(--success); color: var(--success); }
.toast-error { border-left: 4px solid var(--danger); color: var(--danger); }
.hidden { display: none !important; }
@keyframes toast-in {
    from { opacity: 0; transform: translate3d(0, -8px, 0); }
    to { opacity: 1; transform: translate3d(0, 0, 0); }
}

/* Badge */
.badge {
    display: inline-flex; align-items: center;
    padding: 4px 10px;
    border-radius: 999px;
    font-size: 12px; font-weight: 600;
}
.badge-success { background: var(--success-soft); color: var(--success); }
.badge-failed { background: var(--danger-soft); color: var(--danger); }
.badge-idle { background: rgba(255,255,255,0.06); color: var(--text-muted); }

/* Pill / Tag */
.pill {
    display: inline-flex; align-items: center;
    padding: 4px 8px;
    border-radius: 4px;
    background: rgba(255,255,255,0.04);
    color: var(--text-muted);
    font-size: 12px;
    border: 1px solid var(--line);
}
.pill-row { display: flex; gap: 6px; flex-wrap: wrap; }

/* Skeleton */
.skeleton { background: linear-gradient(90deg, rgba(255,255,255,0.04), rgba(255,255,255,0.10), rgba(255,255,255,0.04)); background-size: 200px 100%; animation: shimmer 1.25s linear infinite; border-radius: var(--radius-sm); }
.skeleton-line { display: block; height: 12px; margin-bottom: 10px; }
.skeleton-line:nth-child(1) { width: 44%; }
.skeleton-line:nth-child(2) { width: 82%; }
.skeleton-line:nth-child(3) { width: 68%; }
.skeleton-line:nth-child(4) { width: 90%; }
.skeleton-line:nth-child(5) { width: 58%; margin-bottom: 0; }
@keyframes shimmer { from { background-position: -200px 0; } to { background-position: 200px 0; } }

/* Loading spinner */
.is-loading { position: relative; }
.is-loading::after {
    content: "";
    width: 12px; height: 12px;
    margin-left: 8px;
    display: inline-block;
    vertical-align: -1px;
    border-radius: 50%;
    border: 2px solid currentColor;
    border-right-color: transparent;
    animation: spin 0.7s linear infinite;
}
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

/* Status bar on cards */
.tone-success { border-left: 3px solid var(--success); }
.tone-danger { border-left: 3px solid var(--danger); }
.tone-neutral { border-left: 3px solid var(--primary); }
```

- [ ] **Step 9: 编写总览页面样式**

```css
.hero-console {
    display: flex;
    flex-wrap: wrap;
    gap: 24px;
    padding: 28px;
    background: linear-gradient(135deg, #111627 0%, #0B0F1A 100%);
    border: 1px solid var(--line-strong);
    border-radius: var(--radius-xl);
    margin-bottom: 24px;
}
.hero-main { flex: 1; min-width: 320px; display: grid; gap: 16px; }
.hero-side { min-width: 280px; display: grid; gap: 12px; }
.hero-main h1 { font-size: 28px; line-height: 1.2; letter-spacing: -0.02em; color: var(--text); }
.hero-main p { color: var(--text-soft); line-height: 1.6; }
.hero-eyebrow {
    display: inline-flex; align-items: center; width: fit-content;
    padding: 6px 12px; border-radius: 999px;
    font-size: 12px; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase;
    background: rgba(34, 211, 238, 0.10); color: var(--primary);
}
.hero-metrics { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; }
.hero-metric-card {
    padding: 16px; border-radius: var(--radius-lg);
    border: 1px solid var(--line);
    background: rgba(255,255,255,0.03);
}
.hero-metric-card span { display: block; font-size: 13px; color: var(--text-muted); }
.hero-metric-card strong { display: block; margin-top: 8px; font-size: 16px; color: var(--text); font-family: "Fira Code", monospace; }
.status-card { padding: 16px; border-radius: var(--radius-lg); border: 1px solid var(--line); background: rgba(255,255,255,0.03); }
.status-card strong { display: block; margin: 8px 0 4px; font-size: 16px; }
.status-card p { font-size: 13px; color: var(--text-soft); line-height: 1.5; }
.overview-strip { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 24px; }
.strip-card { padding: 16px 18px; border-radius: var(--radius-lg); border: 1px dashed var(--line-strong); background: rgba(255,255,255,0.02); }
.strip-card span { font-size: 12px; color: var(--text-muted); }
.strip-card strong { display: block; margin-top: 6px; font-size: 14px; color: var(--text-soft); }
.summary-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 24px; }
.summary-card { padding: 20px; }
.summary-card span { display: block; font-size: 13px; color: var(--text-muted); }
.summary-card strong { display: block; margin-top: 14px; font-size: 32px; color: var(--text); letter-spacing: -0.03em; }
.summary-card p { margin-top: 8px; font-size: 13px; color: var(--text-soft); }
.dashboard-grid { display: grid; grid-template-columns: 1.1fr 0.9fr; gap: 16px; margin-bottom: 24px; }
.section-head { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px; }
.section-head > div { display: grid; gap: 6px; }
.section-tag { display: inline-flex; width: fit-content; padding: 4px 10px; border-radius: 999px; font-size: 12px; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; background: rgba(34, 211, 238, 0.10); color: var(--primary); }
.health-bars { display: grid; gap: 12px; }
.health-row { display: grid; gap: 6px; }
.health-label { display: flex; justify-content: space-between; align-items: center; }
.health-label span { font-size: 13px; color: var(--text-soft); }
.health-label strong { font-size: 13px; color: var(--text); font-family: "Fira Code", monospace; }
.health-track { height: 8px; border-radius: 999px; background: rgba(255,255,255,0.06); overflow: hidden; }
.health-fill { height: 100%; border-radius: 999px; }
.fill-success { background: var(--success); }
.fill-failed { background: var(--danger); }
.fill-neutral { background: var(--primary); }
.activity-feed { display: grid; gap: 10px; }
.feed-item { padding: 14px; }
.feed-item strong { display: block; margin-bottom: 4px; font-size: 14px; color: var(--text); }
.feed-item .meta { font-size: 12px; color: var(--text-muted); }
.feed-message { margin-top: 6px; padding: 8px 10px; border-radius: var(--radius-sm); background: rgba(255,255,255,0.03); font-size: 13px; line-height: 1.5; color: var(--text-soft); }
.history-list { display: grid; gap: 10px; }
.history-card { padding: 16px; }
.history-card .task-title { font-size: 15px; font-weight: 600; margin-bottom: 2px; color: var(--text); }
.history-card .meta { font-size: 12px; color: var(--text-muted); }
.history-message { padding: 10px 12px; border-radius: var(--radius-sm); background: rgba(255,255,255,0.03); font-size: 13px; line-height: 1.5; color: var(--text-soft); margin-top: 10px; }
.history-preview { margin-top: 10px; padding: 12px; border-radius: var(--radius-sm); background: rgba(0,0,0,0.2); color: var(--text-soft); font-size: 12px; line-height: 1.6; white-space: pre-wrap; overflow-wrap: anywhere; word-break: break-word; border: 1px solid var(--line); }
.panel { padding: 22px; }
.feature-panel { background: var(--bg-card); border: 1px solid var(--line-strong); border-radius: var(--radius-xl); }
```

- [ ] **Step 10: 编写任务列表页样式**

```css
.task-list-layout { display: grid; gap: 16px; }
.toolbar { display: flex; gap: 12px; align-items: end; flex-wrap: wrap; margin-bottom: 16px; }
.toolbar-field { flex: 1; min-width: 260px; }
.filter-group { display: flex; gap: 6px; flex-wrap: wrap; }
.filter-chip {
    padding: 8px 14px;
    border-radius: 999px;
    border: 1px solid var(--line-strong);
    background: transparent;
    color: var(--text-soft);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
}
.filter-chip:hover { background: rgba(255,255,255,0.04); color: var(--text); }
.filter-chip-active { background: var(--primary-soft); color: var(--primary); border-color: transparent; }
.task-list-body { display: grid; gap: 10px; overflow-y: auto; max-height: calc(100vh - 280px); }
.task-card { padding: 18px; }
.task-card:hover { background: var(--bg-card-hover); }
.task-top { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; }
.task-heading { flex: 1; min-width: 0; }
.task-title { font-size: 16px; font-weight: 600; color: var(--text); margin-bottom: 4px; }
.task-url { font-size: 13px; color: var(--text-muted); word-break: break-all; }
.method-badge {
    display: inline-flex; padding: 4px 8px;
    border-radius: 4px;
    background: rgba(255,255,255,0.06);
    color: var(--text-soft);
    font-size: 11px; font-weight: 700; letter-spacing: 0.04em;
    font-family: "Fira Code", monospace;
}
.meta-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; margin-top: 8px; }
.meta-item { padding: 10px 12px; border-radius: var(--radius-md); background: rgba(255,255,255,0.03); border: 1px solid var(--line); }
.meta-item span { display: block; margin-bottom: 4px; color: var(--text-muted); font-size: 12px; }
.meta-item strong { display: block; color: var(--text-soft); font-size: 13px; }
.task-actions { display: flex; gap: 8px; margin-top: 10px; }
.empty-card { display: grid; gap: 8px; padding: 40px 20px; text-align: center; justify-items: center; }
.empty-title { font-size: 16px; font-weight: 700; color: var(--text-soft); }
```

- [ ] **Step 11: 编写任务编辑器页样式**

```css
.task-editor-page-grid { display: grid; gap: 16px; }
.task-editor-layout { display: grid; gap: 14px; }
.editor-group { padding: 20px; border: 1px solid var(--line); border-radius: var(--radius-lg); background: rgba(255,255,255,0.02); }
.editor-group-head { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 14px; }
.editor-group-head h3 { font-size: 16px; font-weight: 600; color: var(--text); }
.editor-group-head p { font-size: 13px; color: var(--text-muted); margin-top: 4px; }
.form-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.span-2 { grid-column: 1 / -1; }
.schedule-grid { grid-template-columns: repeat(3, 1fr); }
.button-row { display: flex; gap: 8px; flex-wrap: wrap; }
.sticky-actions {
    position: sticky; bottom: 0;
    padding: 14px 0 0;
    background: linear-gradient(180deg, transparent, var(--bg) 38%, var(--bg) 100%);
}
.preview-card { padding: 18px; border: 1px solid var(--line); border-radius: var(--radius-lg); background: rgba(255,255,255,0.02); }
.preview-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.preview { margin: 0; padding: 14px; border-radius: var(--radius-md); background: rgba(0,0,0,0.2); color: var(--text-soft); font-size: 13px; line-height: 1.6; border: 1px solid var(--line); white-space: pre-wrap; overflow-wrap: anywhere; max-height: 220px; overflow-y: auto; font-family: "Fira Code", monospace; }
.toggle-panel { padding: 14px; border-radius: var(--radius-md); background: rgba(255,255,255,0.02); border: 1px solid var(--line); display: flex; gap: 16px; flex-wrap: wrap; }
.compact-grid { margin-bottom: 12px; }
```

- [ ] **Step 12: 编写设置页样式**

```css
.settings-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
.settings-panel { background: var(--bg-card); border: 1px solid var(--line-strong); border-radius: var(--radius-xl); padding: 22px; }
.settings-hero { background: var(--bg-card); border: 1px solid var(--line-strong); border-radius: var(--radius-xl); padding: 24px; display: grid; gap: 6px; margin-bottom: 16px; }
```

- [ ] **Step 13: 编写登录页样式**

```css
.login-body {
    display: grid;
    place-items: center;
    min-height: 100vh;
    padding: 24px;
    background: var(--bg-deepest);
}
.login-shell {
    width: min(1100px, 100%);
    display: grid;
    grid-template-columns: 1.1fr 380px;
    gap: 24px;
    align-items: center;
}
.login-intro { padding: 36px; display: grid; gap: 16px; }
.login-intro h1 { font-size: 32px; color: var(--text); }
.login-intro p { color: var(--text-soft); line-height: 1.6; }
.login-points { display: grid; gap: 12px; }
.login-point { padding: 16px; border-radius: var(--radius-lg); border: 1px solid var(--line); background: rgba(255,255,255,0.02); }
.login-point strong { display: block; margin-bottom: 4px; color: var(--text); font-size: 15px; }
.login-point span { color: var(--text-muted); font-size: 14px; line-height: 1.5; }
.login-card {
    padding: 32px;
    background: var(--bg-card);
    border: 1px solid var(--line-strong);
    border-radius: var(--radius-xl);
    display: grid;
    gap: 16px;
}
.login-card h2 { font-size: 24px; color: var(--text); }
.login-card p { color: var(--text-soft); font-size: 14px; line-height: 1.5; }
.error-box { padding: 10px 14px; border-radius: var(--radius-md); background: var(--danger-soft); color: var(--danger); border: 1px solid var(--danger-soft); font-size: 14px; }
```

- [ ] **Step 14: 编写响应式与无障碍样式**

```css
@media (max-width: 1024px) {
    .summary-grid { grid-template-columns: repeat(2, 1fr); }
    .hero-metrics, .overview-strip, .dashboard-grid, .settings-grid, .task-list-layout { grid-template-columns: 1fr; }
    .sidebar { width: var(--sidebar-collapsed-w); }
    .sidebar .brand-name, .sidebar .brand-sub, .sidebar .sidebar-btn span, .sidebar .sidebar-footer span { display: none; }
    .sidebar-brand { justify-content: center; padding: 16px; }
    .sidebar-btn { justify-content: center; padding: 12px; }
    .sidebar-btn-active::before { left: 0; }
    .sidebar-footer { text-align: center; padding: 16px; }
    .main-content { margin-left: var(--sidebar-collapsed-w); }
    .form-grid { grid-template-columns: 1fr; }
    .schedule-grid { grid-template-columns: repeat(3, 1fr); }
    .login-shell { grid-template-columns: 1fr; }
}
@media (max-width: 640px) {
    .shell { flex-direction: column; }
    .sidebar {
        position: fixed;
        left: 0; top: 0;
        width: 100%; height: 100%;
        transform: translateX(-100%);
        z-index: 200;
    }
    .sidebar.open { transform: translateX(0); }
    .sidebar .brand-name, .sidebar .brand-sub, .sidebar .sidebar-btn span, .sidebar .sidebar-footer span { display: inline; }
    .sidebar-btn { justify-content: flex-start; padding: 12px 14px; }
    .sidebar-brand { justify-content: flex-start; padding: 20px; }
    .sidebar-btn-active::before { left: -12px; }
    .sidebar-footer { text-align: left; padding: 16px 20px; }
    .main-content { margin-left: 0; padding: 16px; }
    .topbar { flex-direction: column; align-items: flex-start; gap: 10px; }
    .toast { top: 16px; right: 16px; left: 16px; min-width: 0; max-width: none; }
    .hero-console, .panel { padding: 20px; }
    h1 { font-size: 24px; }
    .summary-grid, .meta-grid, .hero-metrics, .overview-strip { grid-template-columns: 1fr; }
    .span-2 { grid-column: auto; }
    .sticky-actions { position: static; background: none; padding-top: 0; }
    .login-shell { grid-template-columns: 1fr; }
    .login-body { padding: 16px; }
    .mobile-menu-btn { display: flex; }
}
.mobile-menu-btn {
    display: none;
    background: transparent;
    border: 1px solid var(--line-strong);
    color: var(--text);
    padding: 8px;
    border-radius: var(--radius-md);
}
@media (prefers-reduced-motion: reduce) {
    *, *::before, *::after {
        animation-duration: 0.01ms !important;
        animation-iteration-count: 1 !important;
        transition-duration: 0.01ms !important;
        scroll-behavior: auto !important;
    }
}
```

- [ ] **Step 15: 验证 CSS 编译无语法错误**

Run: 确保文件保存为 UTF-8 编码，检查无明显语法错误

---

### Task 2: HTML - 重构 index.html 为侧边栏布局

**Files:**
- Rewrite: `templates/index.html`

- [ ] **Step 1: 替换顶部导航为侧边栏结构**

将当前顶部导航布局替换为侧边栏 + 主内容区域：

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }}</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Fira+Sans:wght@400;500;600;700&family=Fira+Code:wght@400;500;600;700&family=Noto+Sans+SC:wght@400;500;700&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="/static/app.css?v=20260326d">
</head>
<body>
    <div class="shell">
        <aside class="sidebar" id="sidebar">
            <div class="sidebar-brand">
                <div class="brand-mark">Q</div>
                <div>
                    <div class="brand-name">QianDao</div>
                    <div class="brand-sub">自动签到控制台</div>
                </div>
            </div>
            <nav class="sidebar-nav">
                <button class="sidebar-btn sidebar-btn-active" type="button" id="navOverviewBtn">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
                    <span>总览</span>
                </button>
                <button class="sidebar-btn" type="button" id="navTasksBtn">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
                    <span>任务</span>
                </button>
                <button class="sidebar-btn" type="button" id="navSettingsBtn">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
                    <span>设置</span>
                </button>
            </nav>
            <div class="sidebar-footer">
                <span><span class="status-dot"></span>运行中</span>
                <br>
                <span class="mono">v2.0</span>
            </div>
        </aside>

        <main class="main-content">
            <header class="topbar">
                <div class="topbar-left">
                    <button class="mobile-menu-btn" id="mobileMenuBtn" type="button">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
                    </button>
                    <span class="topbar-title" id="pageTitle">总览</span>
                </div>
                <div class="topbar-actions" id="topbarActions">
                    <!-- 由 JS 动态填充 -->
                </div>
            </header>

            <div id="message" class="toast hidden"></div>

            <!-- 视图容器保持不变，仅更新 class -->
            <section id="overviewView" class="app-view">
                <!-- 内容保持不变 -->
            </section>
            <section id="tasksView" class="app-view hidden">
                <!-- 内容保持不变，task-list-panel 等 class 替换为新样式名 -->
            </section>
            <section id="taskEditorView" class="app-view hidden">
                <!-- 内容保持不变 -->
            </section>
            <section id="settingsView" class="app-view hidden">
                <!-- 内容保持不变 -->
            </section>
        </main>
    </div>
    <script src="/static/app.js?v=20260326d"></script>
</body>
</html>
```

- [ ] **Step 2: 更新总览视图（overviewView）中的 class 名称**

保留 HTML 结构，仅更新 class 名称匹配新的 CSS：
- 删除 `page-orb` 装饰元素（不再需要）
- 删除旧的 `hero-console::after` 背景网格（CSS 已内联）
- `summary-grid` 的 class 保持不变
- `dashboard-grid` → 使用新的 `card` class

- [ ] **Step 3: 更新任务列表视图中的 class 名称**

保留 HTML 结构，更新：
- `filter-chip` → 保留
- `task-list-panel` → 常规 panel 结构
- 任务卡片使用新 class

- [ ] **Step 4: 更新任务编辑器视图中的 class 名称**

保持编辑器结构，使用新 class。

- [ ] **Step 5: 更新设置视图中的 class 名称**

保持设置结构，使用新的 settings-grid class。

- [ ] **Step 6: 确认所有视图使用 card class 保持一致**

确保所有 panel 风格的容器改为 `<section class="card panel">` 或直接使用 `card`。

---

### Task 3: HTML - 重构 login.html

**Files:**
- Modify: `templates/login.html`

- [ ] **Step 1: 更新 login.html 为新设计**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>QianDao V2 登录</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Fira+Sans:wght@400;500;600;700&family=Fira+Code:wght@400;500;600;700&family=Noto+Sans+SC:wght@400;500;700&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="/static/app.css?v=20260326d">
</head>
<body class="login-body">
    <div class="login-shell">
        <section class="login-intro">
            <span class="hero-eyebrow">Secure Access</span>
            <h1>QianDao V2</h1>
            <p>自动签到任务控制台 — 任务管理、秒级调度、Telegram 推送，一站式托管。</p>
            <div class="login-points">
                <div class="login-point">
                    <strong>任务统一管理</strong>
                    <span>保存 Curl 配置、定时执行与批量运行。</span>
                </div>
                <div class="login-point">
                    <strong>秒级调度支持</strong>
                    <span>适合零点签到与需要卡点执行的场景。</span>
                </div>
                <div class="login-point">
                    <strong>默认访问保护</strong>
                    <span>密码门禁适合公网和长期托管部署。</span>
                </div>
            </div>
        </section>

        <form class="login-card" method="post" action="/login">
            <span class="hero-eyebrow">安全登录</span>
            <h2>进入控制台</h2>
            <p>当前控制台已启用访问保护，请输入密码后继续。</p>
            {{ if .Error }}
            <div class="error-box">{{ .Error }}</div>
            {{ end }}
            <label class="field">
                <span>密码</span>
                <input name="password" type="password" required autofocus>
            </label>
            <button class="btn-primary" type="submit" style="justify-content:center">进入控制台</button>
        </form>
    </div>
</body>
</html>
```

---

### Task 4: JS - 导航逻辑更新

**Files:**
- Modify: `static/app.js` (第79-104行改动)

- [ ] **Step 1: 更新 setView 函数以适配侧边栏按钮 ID**

侧边栏按钮 ID 与之前一致（`navOverviewBtn`/`navTasksBtn`/`navSettingsBtn`），但 CSS class 从 `nav-button-active` 改为 `sidebar-btn-active`：

```js
// 替换原有 setView 函数中的按钮切换逻辑
const buttons = {
    overview: $("navOverviewBtn"),
    tasks: $("navTasksBtn"),
    settings: $("navSettingsBtn"),
};

Object.entries(buttons).forEach(([key, element]) => {
    element.classList.toggle("sidebar-btn-active", key === navView);
});
```

- [ ] **Step 2: 更新 setView 函数，更新页面标题**

```js
// 在 setView 末尾添加
const pageTitles = { overview: "总览", tasks: "任务", taskEditor: "编辑任务", settings: "设置" };
$("pageTitle").textContent = pageTitles[view] || "总览";
```

- [ ] **Step 3: 更新 attachEvents 函数，移除旧的顶部按钮绑定**

`$("runAllBtn")`, `$("refreshBtn")`, `$("newTaskBtn")` 这些顶部按钮将被移到 `#topbarActions` 中动态渲染，或在 sidebar 中作为独立按钮。需要将顶栏操作按钮绑定改为动态内容：

```js
function renderTopbarActions() {
    const view = state.currentView;
    let html = "";
    if (view === "overview") {
        html = `<button class="btn-ghost" id="refreshBtn">刷新数据</button>`;
    } else if (view === "tasks") {
        html = `<button class="btn-primary" id="taskPageNewBtn">创建新任务</button>
                <button class="btn-ghost" id="runAllBtn">运行全部</button>`;
    } else if (view === "taskEditor") {
        html = `<button class="btn-ghost" id="taskEditorBackBtn">返回任务列表</button>
                <button class="btn-primary" id="saveTaskBtn">保存任务</button>`;
    } else if (view === "settings") {
        // 不需要额外操作
    }
    $("topbarActions").innerHTML = html;
    // 重新绑定事件
    rebindTopbarEvents();
}
```

- [ ] **Step 4: 更新 mobile menu 按钮绑定**

```js
$("mobileMenuBtn").addEventListener("click", () => {
    document.getElementById("sidebar").classList.toggle("open");
});
// 点击主内容区关闭侧边栏
document.querySelector(".main-content").addEventListener("click", () => {
    document.getElementById("sidebar").classList.remove("open");
});
```

- [ ] **Step 5: 更新所有按钮 class 引用**

JS 中按钮 class 名称更新（在 `getTaskActionMarkup` 等函数中）：
- `.primary` → `.btn-primary`
- `.secondary` → `.btn-secondary`
- `.ghost` → `.btn-ghost`
- `.danger` → `.btn-danger`

- [ ] **Step 6: 验证所有功能正常**

启动项目，检查：
- 导航切换正常
- 总览数据加载正常
- 任务列表正常
- 编辑器正常
- 设置页面正常
- 登录页面正常
- 响应式布局正常
