from flask import Flask, render_template, request, jsonify, redirect, url_for, session
import requests
import json
import re
import os
import hashlib

app = Flask(__name__)
app.secret_key = 'your-secret-key'

# 配置文件路径
CONFIG_FILE = 'signin_configs.json'
NOTIFY_CONFIG_FILE = 'notify_config.json'
SCHEDULE_CONFIG_FILE = 'schedule_config.json'
PASSWORD_CONFIG_FILE = 'password_config.json'

# 加载签到配置
def load_configs():
    if os.path.exists(CONFIG_FILE):
        with open(CONFIG_FILE, 'r', encoding='utf-8') as f:
            return json.load(f)
    return []

# 保存签到配置
def save_configs(configs):
    with open(CONFIG_FILE, 'w', encoding='utf-8') as f:
        json.dump(configs, f, ensure_ascii=False, indent=2)

# 加载通知配置
def load_notify_config():
    if os.path.exists(NOTIFY_CONFIG_FILE):
        with open(NOTIFY_CONFIG_FILE, 'r', encoding='utf-8') as f:
            return json.load(f)
    return {'tg_bot_token': '', 'tg_chat_id': ''}

# 保存通知配置
def save_notify_config(config):
    with open(NOTIFY_CONFIG_FILE, 'w', encoding='utf-8') as f:
        json.dump(config, f, ensure_ascii=False, indent=2)

# 发送Telegram通知
def send_telegram_notification(message):
    notify_config = load_notify_config()
    bot_token = notify_config.get('tg_bot_token')
    chat_id = notify_config.get('tg_chat_id')
    
    if not bot_token or not chat_id:
        return False
    
    try:
        url = f"https://api.telegram.org/bot{bot_token}/sendMessage"
        data = {
            'chat_id': chat_id,
            'text': message,
            'parse_mode': 'Markdown'
        }
        response = requests.post(url, data=data, timeout=10)
        response.raise_for_status()
        return True
    except Exception as e:
        print(f"发送TG通知失败: {e}")
        return False

# 检查并执行定时任务
def check_schedule():
    import datetime
    now = datetime.datetime.now()
    current_hour = now.hour
    current_minute = now.minute
    
    configs = load_configs()
    schedule_config = load_schedule_config()
    
    for config in configs:
        # 检查任务是否有单独的定时设置
        task_schedule = config.get('schedule', {})
        task_enabled = task_schedule.get('enabled', False)
        
        if task_enabled:
            # 使用任务单独的定时设置
            task_hour = task_schedule.get('hour', 0)
            task_minute = task_schedule.get('minute', 0)
            
            if current_hour == task_hour and current_minute == task_minute:
                print(f"执行定时任务: {config.get('name', '未知')}")
                run_signin(config)
        elif schedule_config.get('enabled', False):
            # 使用通用的定时设置
            global_hour = schedule_config.get('hour', 0)
            global_minute = schedule_config.get('minute', 0)
            
            if current_hour == global_hour and current_minute == global_minute:
                print(f"执行通用定时任务: {config.get('name', '未知')}")
                run_signin(config)

# 加载定时任务配置
def load_schedule_config():
    if os.path.exists(SCHEDULE_CONFIG_FILE):
        with open(SCHEDULE_CONFIG_FILE, 'r', encoding='utf-8') as f:
            return json.load(f)
    return {'enabled': False, 'hour': 0, 'minute': 0}

# 保存定时任务配置
def save_schedule_config(config):
    with open(SCHEDULE_CONFIG_FILE, 'w', encoding='utf-8') as f:
        json.dump(config, f, ensure_ascii=False, indent=2)

# 加载密码配置
def load_password_config():
    if os.path.exists(PASSWORD_CONFIG_FILE):
        with open(PASSWORD_CONFIG_FILE, 'r', encoding='utf-8') as f:
            return json.load(f)
    # 默认无密码
    return {'enabled': False, 'password_hash': ''}

# 保存密码配置
def save_password_config(config):
    with open(PASSWORD_CONFIG_FILE, 'w', encoding='utf-8') as f:
        json.dump(config, f, ensure_ascii=False, indent=2)

# 密码哈希函数
def hash_password(password):
    return hashlib.sha256(password.encode()).hexdigest()

# 密码验证函数
def verify_password(password):
    config = load_password_config()
    if not config['enabled']:
        return True  # 未启用密码验证
    return hash_password(password) == config['password_hash']

# 登录装饰器
def login_required(f):
    def decorated_function(*args, **kwargs):
        password_config = load_password_config()
        if password_config['enabled'] and 'logged_in' not in session:
            return redirect(url_for('login'))
        return f(*args, **kwargs)
    decorated_function.__name__ = f.__name__
    return decorated_function

# 解析Curl命令
def parse_curl(curl_command):
    config = {
        'method': 'GET',
        'headers': {},
        'data': None,
        'schedule': {
            'enabled': False,
            'hour': 0,
            'minute': 0
        }
    }
    
    # 提取URL
    url_match = re.search(r"curl\s+'([^']+)'", curl_command)
    if not url_match:
        raise Exception('无法提取URL')
    config['url'] = url_match.group(1)
    
    # 提取方法
    method_match = re.search(r"-X\s+(\w+)", curl_command)
    if method_match:
        config['method'] = method_match.group(1)
    
    # 提取headers
    header_matches = re.findall(r"-H\s+'([^']+)'", curl_command)
    for header in header_matches:
        if ': ' in header:
            key, value = header.split(': ', 1)
            config['headers'][key] = value
    
    # 提取cookie
    cookie_match = re.search(r"-b\s+'([^']+)'", curl_command)
    if cookie_match:
        cookie = cookie_match.group(1)
        config['headers']['Cookie'] = cookie
    
    # 提取data
    data_match = re.search(r"-d\s+'([^']+)'", curl_command)
    if data_match:
        config['data'] = data_match.group(1)
    
    return config

# 运行签到
def run_signin(config):
    try:
        method = config['method']
        url = config['url']
        headers = config['headers']
        data = config['data']
        
        if method == 'GET':
            response = requests.get(url, headers=headers, timeout=30)
        elif method == 'POST':
            response = requests.post(url, headers=headers, data=data, timeout=30)
        else:
            response = requests.request(method, url, headers=headers, data=data, timeout=30)
        
        # 获取返回内容
        content = response.text
        status_code = response.status_code
        
        # 检查HTTP状态码是否为200
        if status_code != 200:
            # 发送失败通知（状态码不是200）
            message = f"❌ **签到失败**\n\n**网站:** {config.get('name', '未知')}\n**URL:** {url}\n**错误:** HTTP {status_code}\n**详情:** 状态码不是200"
            send_telegram_notification(message)
            error_detail = content[:100] + '...' if content else '无响应内容'
            return {'success': False, 'error': f'HTTP {status_code} - {error_detail}', 'status_code': status_code, 'content': content}
        
        # 定义错误关键词列表（包括简体中文、繁体中文和英文）
        error_keywords = [
            # 简体中文
            '错误', '失败', '无效', '未找到', '拒绝', '异常', '错误码', 'error', 'fail', 'invalid', 'not found', 'denied', 'exception', 'error code',
            # 繁体中文
            '錯誤', '失敗', '無效', '未找到', '拒絕', '異常', '錯誤碼'
        ]
        
        # 定义成功关键词列表
        success_keywords = [
            # 简体中文
            '成功', '已签到', '签到成功', 'success', 'signed', 'sign in success', 'checked in', 'check in success',
            # 繁体中文
            '成功', '已簽到', '簽到成功'
        ]
        
        # 检查是否包含错误关键词
        has_error = any(keyword in content.lower() for keyword in error_keywords)
        
        # 检查是否包含成功关键词
        has_success = any(keyword in content.lower() for keyword in success_keywords)
        
        # 判断最终结果
        if has_error and not has_success:
            # 包含错误关键词且不包含成功关键词，视为失败
            message = f"❌ **签到失败**\n\n**网站:** {config.get('name', '未知')}\n**URL:** {url}\n**状态码:** {status_code}\n**错误:** 返回数据中包含错误信息"
            send_telegram_notification(message)
            error_detail = content[:100] + '...' if content else '无响应内容'
            return {'success': False, 'error': f'返回数据中包含错误信息 - {error_detail}', 'status_code': status_code, 'content': content}
        else:
            # 状态码200且不包含错误信息，视为成功
            message = f"✅ **签到成功**\n\n**网站:** {config.get('name', '未知')}\n**URL:** {url}\n**状态码:** {status_code}\n**状态:** 成功"
            send_telegram_notification(message)
            return {'success': True, 'content': content, 'status_code': status_code}
    except requests.exceptions.HTTPError as e:
            # 发送失败通知（HTTP错误）
            status_code = e.response.status_code if e.response else None
            message = f"❌ **签到失败**\n\n**网站:** {config.get('name', '未知')}\n**URL:** {url}\n**错误:** HTTP {status_code if status_code else '未知状态码'}\n**详情:** {str(e)}"
            send_telegram_notification(message)
            error_detail = e.response.text[:100] + '...' if e.response else '无响应内容'
            return {'success': False, 'error': f'{str(e)} - {error_detail}', 'status_code': status_code, 'content': e.response.text if e.response else ''}
    except Exception as e:
        # 发送失败通知（其他错误）
        message = f"❌ **签到失败**\n\n**网站:** {config.get('name', '未知')}\n**URL:** {url}\n**错误:** {str(e)}"
        send_telegram_notification(message)
        return {'success': False, 'error': str(e), 'status_code': None, 'content': ''}

@app.route('/login', methods=['GET', 'POST'])
def login():
    if request.method == 'POST':
        password = request.form['password']
        if verify_password(password):
            session['logged_in'] = True
            return redirect(url_for('index'))
        else:
            return render_template('login.html', error='密码错误')
    # 检查是否已启用密码验证
    password_config = load_password_config()
    if not password_config['enabled']:
        return redirect(url_for('index'))
    return render_template('login.html')

@app.route('/logout')
def logout():
    session.pop('logged_in', None)
    return redirect(url_for('login'))

@app.route('/')
@login_required
def index():
    configs = load_configs()
    # 读取HTML文件并嵌入配置数据
    if os.path.exists('templates/index.html'):
        with open('templates/index.html', 'r', encoding='utf-8') as f:
            content = f.read()
        # 嵌入配置数据
        content = content.replace('</body>', f'<script id="configs-data" type="application/json">{json.dumps(configs, ensure_ascii=False)}</script>\n</body>')
        return content
    return '模板文件不存在'

@app.route('/parse', methods=['POST'])
@login_required
def parse():
    curl_command = request.form['curl']
    site_name = request.form['name']
    request_method = request.form.get('method', 'GET')  # 获取用户选择的请求方法，默认为GET
    task_enabled = request.form.get('taskEnabled', 'false') == 'true'
    task_hour = int(request.form.get('taskHour', '8'))
    task_minute = int(request.form.get('taskMinute', '0'))
    
    try:
        config = parse_curl(curl_command)
        config['name'] = site_name
        config['method'] = request_method  # 使用用户选择的请求方法
        config['schedule'] = {
            'enabled': task_enabled,
            'hour': task_hour,
            'minute': task_minute
        }
        return jsonify({'success': True, 'config': config})
    except Exception as e:
        return jsonify({'success': False, 'error': str(e)})

@app.route('/save', methods=['POST'])
@login_required
def save():
    config = request.json
    configs = load_configs()
    configs.append(config)
    save_configs(configs)
    return jsonify({'success': True})

@app.route('/run/<int:index>', methods=['POST'])
@login_required
def run(index):
    configs = load_configs()
    if 0 <= index < len(configs):
        result = run_signin(configs[index])
        return jsonify(result)
    return jsonify({'success': False, 'error': '配置不存在'})

@app.route('/run-all', methods=['POST'])
@login_required
def run_all():
    configs = load_configs()
    results = []
    for i, config in enumerate(configs):
        result = run_signin(config)
        results.append({'index': i, 'name': config['name'], 'result': result})
    return jsonify({'success': True, 'results': results})

@app.route('/delete/<int:index>', methods=['POST'])
@login_required
def delete(index):
    configs = load_configs()
    if 0 <= index < len(configs):
        configs.pop(index)
        save_configs(configs)
        return jsonify({'success': True})
    return jsonify({'success': False, 'error': '配置不存在'})

@app.route('/edit/<int:index>', methods=['GET'])
@login_required
def edit(index):
    configs = load_configs()
    if 0 <= index < len(configs):
        config = configs[index]
        # 生成curl命令
        curl = f"curl '{config['url']}'"
        if config['method'] != 'GET':
            curl += f" \\\n  -X {config['method']}"
        for key, value in config['headers'].items():
            curl += f" \\\n  -H '{key}: {value}'"
        if config['data']:
            curl += f" \\\n  -d '{config['data']}'"
        config['curl'] = curl
        # 确保schedule字段存在
        if 'schedule' not in config:
            config['schedule'] = {
                'enabled': False,
                'hour': 8,
                'minute': 0
            }
        return jsonify({'success': True, 'config': config, 'index': index})
    return jsonify({'success': False, 'error': '配置不存在'})

@app.route('/update/<int:index>', methods=['POST'])
@login_required
def update(index):
    config = request.json
    configs = load_configs()
    if 0 <= index < len(configs):
        configs[index] = config
        save_configs(configs)
        return jsonify({'success': True})
    return jsonify({'success': False, 'error': '配置不存在'})

@app.route('/notify/config', methods=['GET'])
@login_required
def get_notify_config():
    config = load_notify_config()
    return jsonify(config)

@app.route('/notify/save', methods=['POST'])
@login_required
def save_notify_config_route():
    config = request.json
    save_notify_config(config)
    return jsonify({'success': True})

@app.route('/notify/test', methods=['POST'])
@login_required
def test_notify():
    success = send_telegram_notification("📢 **测试通知**\n\n这是一条测试消息，说明Telegram通知配置成功！")
    return jsonify({'success': success})

@app.route('/schedule/config', methods=['GET'])
@login_required
def get_schedule_config():
    config = load_schedule_config()
    return jsonify(config)

@app.route('/schedule/save', methods=['POST'])
@login_required
def save_schedule_config_route():
    config = request.json
    save_schedule_config(config)
    return jsonify({'success': True})

@app.route('/schedule/check', methods=['POST'])
@login_required
def check_schedule_route():
    check_schedule()
    return jsonify({'success': True})

# 密码配置路由
@app.route('/password/config', methods=['GET'])
@login_required
def get_password_config():
    config = load_password_config()
    return jsonify(config)

@app.route('/password/save', methods=['POST'])
@login_required
def save_password_config_route():
    config = request.json
    if config.get('password'):
        config['password_hash'] = hash_password(config['password'])
        del config['password']  # 删除明文密码
    save_password_config(config)
    return jsonify({'success': True})

import threading
import time

# 后台线程检查定时任务
def schedule_thread():
    while True:
        check_schedule()
        time.sleep(60)  # 每分钟检查一次

if __name__ == '__main__':
    # 确保templates目录存在
    if not os.path.exists('templates'):
        os.makedirs('templates')
    
    # 复制index.html到templates目录
    if not os.path.exists('templates/index.html') and os.path.exists('index.html'):
        with open('index.html', 'r', encoding='utf-8') as f:
            content = f.read()
        with open('templates/index.html', 'w', encoding='utf-8') as f:
            f.write(content)
    
    # 启动定时任务检查线程
    thread = threading.Thread(target=schedule_thread, daemon=True)
    thread.start()
    
    import os
    debug_mode = os.environ.get('FLASK_ENV') != 'production'
    app.run(host='0.0.0.0', port=5000, debug=debug_mode)
