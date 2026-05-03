// ==================== 全局变量 ====================
let ws = null;
let currentChatId = 0;
let currentUserId = 0;
let userInfo = null;  // 当前登录用户信息

// 未读消息管理
let unreadMessageCount = 0;
let unreadFriendRequestCount = 0;
let pendingFriendRequest = null;
let toastTimeout = null;

// 未读消息详情（按会话统计）
const unreadMessages = {};  // friendId -> count

// 已确认消息集合（用于防止重复接收）
const confirmedMessages = new Set();  // msgId -> true

// WebSocket连接状态追踪
let wsReconnectTimer = null;
let wsReconnectAttempts = 0;
const MAX_WS_RECONNECT = 5;
let wsConnecting = false;  // 防止重复连接

// 心跳配置
const HEARTBEAT_INTERVAL = 15000; // 15秒发送一次心跳（减小心跳窗口）
const HEARTBEAT_TIMEOUT = 45000;  // 45秒未收到响应认为断开（3倍安全边际）
let heartbeatTimer = null;
let lastPongTime = 0;

// 定时刷新好友在线状态
let friendStatusTimer = null;

// 缓存数据
const cache = {
    friendList: null,
    messageList: null,
    chatHistory: {}  // 聊天历史缓存（内存）
};

// 在线状态缓存（通过WebSocket实时更新）
const onlineStatusCache = new Map();

// localStorage 键名
const STORAGE_KEYS = {
    CHAT_HISTORY: 'littletalk_chat_history_',  // 聊天历史，按 friend_id 存储
    RECENT_CHATS: 'littletalk_recent_chats',    // 最近联系人列表
    USER_INFO: 'littletalk_user_',              // 用户信息
    CHAT_BG: 'littletalk_chat_bg'               // 聊天背景设置
};

// 预设聊天背景列表
const CHAT_BG_SOLID = [
    { id: 'solid-default', name: '默认灰', color: '#f5f7fa' },
    { id: 'solid-blue', name: '天空蓝', color: '#e3f2fd' },
    { id: 'solid-green', name: '抹茶绿', color: '#e8f5e9' },
    { id: 'solid-pink', name: '樱花粉', color: '#fce4ec' },
    { id: 'solid-purple', name: '薰衣草', color: '#f3e5f5' },
    { id: 'solid-orange', name: '暖阳橙', color: '#fff3e0' },
    { id: 'solid-teal', name: '薄荷绿', color: '#e0f2f1' },
    { id: 'solid-peach', name: '蜜桃粉', color: '#fde8d8' },
    { id: 'solid-lavender', name: '淡紫蓝', color: '#e8eaf6' },
    { id: 'solid-mint', name: '清新薄荷', color: '#e8f5e9' },
];

const CHAT_BG_GRADIENT = [
    { id: 'grad-aurora', name: '极光', gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' },
    { id: 'grad-sunrise', name: '日出', gradient: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' },
    { id: 'grad-forest', name: '森林', gradient: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)' },
    { id: 'grad-ocean', name: '海洋', gradient: 'linear-gradient(135deg, #2193b0 0%, #6dd5ed 100%)' },
    { id: 'grad-sunset', name: '落日', gradient: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)' },
    { id: 'grad-night', name: '夜空', gradient: 'linear-gradient(135deg, #0f2027 0%, #203a43 50%, #2c5364 100%)' },
    { id: 'grad-dream', name: '梦幻', gradient: 'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)' },
    { id: 'grad-fire', name: '烈焰', gradient: 'linear-gradient(135deg, #ff4b1f 0%, #ff9068 100%)' },
    { id: 'grad-neon', name: '霓虹', gradient: 'linear-gradient(135deg, #ff0844 0%, #ffb199 100%)' },
    { id: 'grad-warm', name: '暖阳', gradient: 'linear-gradient(135deg, #f6d365 0%, #fda085 100%)' },
];

// 默认背景设置
const DEFAULT_BG = CHAT_BG_SOLID[0];

// 用户信息缓存
const userCache = {
    data: {},

    // 添加或更新用户
    set(userId, info) {
        this.data[userId] = {
            id: userId,
            username: info.username || info.name || `用户${userId}`,
            avatar: info.avatar || '',
            ...info
        };
    },

    // 批量添加用户
    setMulti(users) {
        users.forEach(user => {
            this.set(user.id, user);
        });
    },

    // 获取用户信息
    get(userId) {
        return this.data[userId] || null;
    },

    // 获取用户名
    getUsername(userId) {
        const user = this.data[userId];
        return user ? user.username : `用户${userId}`;
    },

    // 获取用户头像
    getAvatar(userId) {
        const user = this.data[userId];
        return user ? user.avatar : '';
    },

    // 从服务器获取用户信息（仅获取缓存中没有的）
    async fetchMissing(userIds) {
        const missingIds = userIds.filter(id => !this.data[id]);
        if (missingIds.length === 0) return;

        try {
            const { data } = await axios.get('/api/usersinfo', {
                params: { ids: missingIds.join(',') }
            });
            if (data.code === 0 && data.data) {
                this.setMulti(data.data);
            }
        } catch (error) {
            console.error('获取用户信息失败:', error);
        }
    }
};

// ==================== 基础配置 ====================
const BASE_URL = window.location.protocol + '//' + window.location.host;
axios.defaults.baseURL = BASE_URL;
axios.defaults.withCredentials = true;

// WebSocket URL配置
function getWsUrl() {
    const protocol = BASE_URL.startsWith('https') ? 'wss' : 'ws';
    const host = BASE_URL.replace(/^https?:\/\//, '');
    return `${protocol}://${host}/api/ws`;
}

// 页面加载时自动设置Authorization头并检查登录状态
(function initAuthHeader() {
    const token = getCookie('token');
    if (token) {
        axios.defaults.headers.common['Authorization'] = 'Bearer ' + token;
        // Token存在，自动登录并连接WebSocket
        autoLoginWithToken(token);
    }
})();

// 使用token自动登录
async function autoLoginWithToken(token) {
    try {
        const { data } = await axios.get('/api/selfuserinfo');
        if (data.code === 0) {
            userInfo = data.data;
            currentUserId = userInfo.id;
            userCache.set(userInfo.id, {
                username: userInfo.username,
                avatar: userInfo.avatar
            });

            // 显示主页面
            showMainPage();
            getSelfUserInfo();
            getMessageList();
            getFriendList();
            getFriendRequestList();
            loadUnreadCounts();  // 加载未读消息数量

            // 连接WebSocket
            connectWebSocket();
            startFriendStatusRefresh();
        } else {
            // Token无效，清除
            deleteCookie('token');
        }
    } catch (error) {
        console.log('自动登录失败:', error);
        deleteCookie('token');
    }
}

// 加载未读消息数量
async function loadUnreadCounts() {
    try {
        const { data } = await axios.get('/api/unreadcount');
        if (data.code === 0) {
            unreadMessageCount = data.data.total || 0;
            updateMessageBadge();
        }
    } catch (error) {
        console.log('获取未读消息数失败:', error);
    }
}

// ==================== 本地存储管理 ====================

// 获取聊天历史（从localStorage）
function getLocalChatHistory(friendId) {
    const key = STORAGE_KEYS.CHAT_HISTORY + friendId;
    const data = localStorage.getItem(key);
    return data ? JSON.parse(data) : [];
}

// 保存聊天历史到localStorage
function saveLocalChatHistory(friendId, messages) {
    const key = STORAGE_KEYS.CHAT_HISTORY + friendId;
    // 最多保存100条
    const trimmed = messages.slice(-100);
    localStorage.setItem(key, JSON.stringify(trimmed));
}

// 添加单条消息到本地历史
function addMessageToLocalHistory(friendId, message) {
    const history = getLocalChatHistory(friendId);
    history.push(message);
    saveLocalChatHistory(friendId, history);
}

// 更新本地历史中的消息状态
function updateMessageInLocalHistory(friendId, msgId, status) {
    const history = getLocalChatHistory(friendId);
    const msg = history.find(m => m.msg_id === msgId);
    if (msg) {
        msg.status = status;
        saveLocalChatHistory(friendId, history);
    }
}

// 获取最近联系人列表
function getLocalRecentChats() {
    const data = localStorage.getItem(STORAGE_KEYS.RECENT_CHATS);
    return data ? JSON.parse(data) : [];
}

// 保存最近联系人列表
function saveLocalRecentChats(chats) {
    localStorage.setItem(STORAGE_KEYS.RECENT_CHATS, JSON.stringify(chats));
}

// 更新最近联系人的最后一条消息
function updateLocalRecentChat(friendId, friendName, friendAvatar, message, sendTime) {
    let recentChats = getLocalRecentChats();

    // 查找是否已存在
    const index = recentChats.findIndex(c => c.friend_id === friendId);

    const chatInfo = {
        friend_id: friendId,
        friend_name: friendName,
        friend_avatar: friendAvatar,
        last_message: message,
        send_time: sendTime
    };

    if (index >= 0) {
        // 更新现有
        recentChats.splice(index, 1);
    }

    // 添加到最前面
    recentChats.unshift(chatInfo);

    // 最多保留20个最近联系人
    recentChats = recentChats.slice(0, 20);

    saveLocalRecentChats(recentChats);
}

// ==================== Cookie辅助函数 ====================
function setCookie(name, value, days = 7) {
    const expires = new Date();
    expires.setTime(expires.getTime() + days * 24 * 60 * 60 * 1000);
    document.cookie = name + '=' + encodeURIComponent(value) + ';expires=' + expires.toUTCString() + ';path=/';
}

function getCookie(name) {
    const nameEQ = name + '=';
    const ca = document.cookie.split(';');
    for (let i = 0; i < ca.length; i++) {
        let c = ca[i];
        while (c.charAt(0) === ' ') c = c.substring(1, c.length);
        if (c.indexOf(nameEQ) === 0) return decodeURIComponent(c.substring(nameEQ.length, c.length));
    }
    return null;
}

function deleteCookie(name) {
    document.cookie = name + '=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/';
}

function getToken() {
    return getCookie('token');
}

// ==================== 页面切换 ====================
function showLoginForm() {
    document.getElementById('register-form').classList.add('hidden');
    document.querySelector('.login-title').classList.remove('hidden');
    document.querySelector('.login-logo').classList.remove('hidden');
    document.querySelector('button[onclick="loginTest()"]').classList.remove('hidden');
    document.getElementById('login-response').classList.remove('hidden');
    document.querySelector('.register-link:not(:last-child)').classList.remove('hidden');
}

function showRegisterForm() {
    document.getElementById('register-form').classList.remove('hidden');
    document.querySelector('.login-title').classList.add('hidden');
    document.querySelector('.login-logo').classList.add('hidden');
    document.querySelector('button[onclick="loginTest()"]').classList.add('hidden');
    document.getElementById('login-response').classList.add('hidden');
    document.querySelector('.register-link:not(:last-child)').classList.add('hidden');
}

function showMainPage() {
    document.getElementById('login-page').classList.add('hidden');
    document.getElementById('main-page').classList.remove('hidden');
}

function showAddFriend() {
    document.getElementById('main-page').classList.add('hidden');
    document.getElementById('add-friend-page').classList.remove('hidden');
}

function showSettings() {
    alert('设置功能开发中...');
}

function backToMain() {
    document.getElementById('chat-page').classList.add('hidden');
    document.getElementById('add-friend-page').classList.add('hidden');
    document.getElementById('main-page').classList.remove('hidden');
}

function openChat(id, name) {
    currentChatId = id;
    document.getElementById('chat-name').textContent = name;
    document.getElementById('main-page').classList.add('hidden');
    document.getElementById('chat-page').classList.remove('hidden');
    loadChatHistory(id);
}

function switchTab(tabId, event) {
    // 隐藏所有tab内容
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.add('hidden');
    });

    // 显示选中的tab内容
    document.getElementById(tabId).classList.remove('hidden');

    // 更新tab状态
    document.querySelectorAll('.tab-item').forEach(item => {
        item.classList.remove('active');
    });

    if (event) {
        event.target.closest('.tab-item').classList.add('active');
    }

    // Tab特定逻辑
    if (tabId === 'friend-tab') {
        // 加载好友请求列表
        getFriendRequestListForUI();
        // 切换到好友tab时清零好友申请未读数
        unreadFriendRequestCount = 0;
        updateFriendRequestBadge();
    }

    if (tabId === 'message-tab') {
        unreadMessageCount = 0;
        updateMessageBadge();
    }
}

// ==================== 未读消息红点 ====================
function updateMessageBadge() {
    const badge = document.getElementById('message-unread-badge');
    if (unreadMessageCount > 0) {
        badge.textContent = unreadMessageCount > 99 ? '99+' : unreadMessageCount;
        badge.classList.remove('hidden');
    } else {
        badge.classList.add('hidden');
    }
}

function updateFriendRequestBadge() {
    const badge = document.getElementById('friend-request-badge');
    const countEl = document.getElementById('friend-request-count');
    if (unreadFriendRequestCount > 0) {
        badge.textContent = unreadFriendRequestCount > 99 ? '99+' : unreadFriendRequestCount;
        badge.classList.remove('hidden');
        if (countEl) {
            countEl.textContent = unreadFriendRequestCount;
            countEl.classList.remove('hidden');
        }
    } else {
        badge.classList.add('hidden');
        if (countEl) {
            countEl.textContent = '0';
            countEl.classList.add('hidden');
        }
    }
}

// ==================== 好友请求弹窗 ====================
function showFriendRequestModal() {
    const modal = document.getElementById('friend-request-modal');
    modal.classList.remove('hidden');
    // 打开弹窗时清零未读数
    unreadFriendRequestCount = 0;
    updateFriendRequestBadge();
    getFriendRequestListForUI();
}

function closeFriendRequestModal() {
    const modal = document.getElementById('friend-request-modal');
    modal.classList.add('hidden');
}

// 点击弹窗背景关闭
function handleModalBackgroundClick(event) {
    if (event.target.classList.contains('modal')) {
        closeFriendRequestModal();
    }
}

// ==================== 通知弹窗 ====================
function showFriendRequestToast(fromUsername) {
    pendingFriendRequest = {
        fromUsername: fromUsername,
        fromId: currentChatId
    };

    const toast = document.getElementById('friend-request-toast');
    const body = document.getElementById('friend-request-toast-body');
    body.textContent = fromUsername + ' 想要添加你为好友';
    toast.classList.remove('hidden');

    // 5秒后自动隐藏
    if (toastTimeout) clearTimeout(toastTimeout);
    toastTimeout = setTimeout(() => {
        toast.classList.add('hidden');
    }, 5000);

    playNotificationSound();
    unreadFriendRequestCount++;
    updateFriendRequestBadge();
}

function showMessageToast(fromId, content) {
    const toast = document.getElementById('new-message-toast');
    const body = document.getElementById('new-message-toast-body');

    // 从缓存获取用户信息
    const userInfo = userCache.get(fromId);
    const username = userInfo ? userInfo.username : `用户${fromId}`;
    const avatar = userInfo ? userInfo.avatar : '';
    const initial = username.charAt(0).toUpperCase();

    // 头像HTML
    const avatarHtml = avatar
        ? `<img src="${avatar}" alt="${initial}" onerror="this.style.display='none'; this.nextSibling.style.display='flex';"><span style="display:none;">${initial}</span>`
        : `<span>${initial}</span>`;

    body.innerHTML = `
        <div class="msg-item">
            <div class="msg-avatar">${avatarHtml}</div>
            <div class="msg-info">
                <div class="msg-name">${username}</div>
                <div class="msg-text">${escapeHtml(content)}</div>
            </div>
        </div>
    `;

    toast.classList.remove('hidden');

    window._pendingChatFromId = fromId;
    window._pendingChatFromName = username;

    if (toastTimeout) clearTimeout(toastTimeout);
    toastTimeout = setTimeout(() => {
        toast.classList.add('hidden');
    }, 5000);

    playNotificationSound();
    unreadMessageCount++;
    updateMessageBadge();
}

function playNotificationSound() {
    try {
        const audioContext = new (window.AudioContext || window.webkitAudioContext)();
        const oscillator = audioContext.createOscillator();
        const gainNode = audioContext.createGain();

        oscillator.connect(gainNode);
        gainNode.connect(audioContext.destination);

        oscillator.frequency.value = 800;
        oscillator.type = 'sine';
        gainNode.gain.value = 0.1;

        oscillator.start();
        oscillator.stop(audioContext.currentTime + 0.1);
    } catch (e) {
        console.log('提示音播放失败:', e);
    }
}

async function handleFriendRequestToast(action) {
    const toast = document.getElementById('friend-request-toast');
    toast.classList.add('hidden');

    if (!pendingFriendRequest) return;

    const fromId = pendingFriendRequest.fromId;

    try {
        if (action === 'accept') {
            await axios.post('/api/friend/respond', {
                from_id: fromId,
                action: 'accept'
            });
            alert('已接受好友请求');
            getFriendList();
        } else {
            await axios.post('/api/friend/respond', {
                from_id: fromId,
                action: 'reject'
            });
        }
        getFriendRequestList();
    } catch (error) {
        console.error('处理好友请求失败:', error);
        alert('处理失败');
    }

    pendingFriendRequest = null;
}

function openChatFromToast() {
    const toast = document.getElementById('new-message-toast');
    toast.classList.add('hidden');

    if (window._pendingChatFromId) {
        switchTab('message-tab');
        openChat(window._pendingChatFromId, window._pendingChatFromName);
    }
}

// ==================== 登录注册 ====================

// 清除本地用户缓存（切换账号时调用）
function clearUserLocalCache() {
    // 清除聊天历史
    const keysToRemove = [];
    for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key && key.startsWith('littletalk_')) {
            keysToRemove.push(key);
        }
    }
    keysToRemove.forEach(key => localStorage.removeItem(key));
    
    // 清除内存中的缓存
    onlineStatusCache.clear();
    confirmedMessages.clear();
    cache.chatHistory = {};
    // 清除未读消息
    Object.keys(unreadMessages).forEach(key => delete unreadMessages[key]);
    unreadMessageCount = 0;
    updateMessageBadge();
    console.log('本地用户缓存已清除');
}

async function loginTest() {
    const username = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;
    const responseDiv = document.getElementById('login-response');

    try {
        responseDiv.textContent = '登录中...';
        const { data } = await axios.post('/login', { username, password });

        if (data.code === 0) {
            // 密码登录：清除旧的用户缓存，避免多账号数据混乱
            clearUserLocalCache();
            
            currentUserId = data.data.user_id;
            setCookie('token', data.data.token, 7);
            axios.defaults.headers.common['Authorization'] = 'Bearer ' + data.data.token;

            alert('登录成功！');
            showMainPage();
            getSelfUserInfo();
            getMessageList();
            getFriendList();
            getFriendRequestList();
            connectWebSocket();
            startFriendStatusRefresh();
        } else {
            responseDiv.textContent = '登录失败: ' + (data.message || '未知错误');
        }
    } catch (error) {
        const errorMessage = error.response?.data?.message || error.message || '网络错误';
        responseDiv.textContent = '错误: ' + errorMessage;
    }
}

async function registerTest() {
    const username = document.getElementById('register-username').value;
    const password = document.getElementById('register-password').value;
    const sex = parseInt(document.getElementById('register-sex').value);
    const birthday = document.getElementById('register-birthday').value;
    const responseDiv = document.getElementById('register-response');

    try {
        responseDiv.textContent = '注册中...';
        const { data } = await axios.post('/creatuser', { username, password, sex, birthday });

        if (data.code === 0) {
            alert('注册成功！请登录');
            showLoginForm();
        } else {
            responseDiv.textContent = '注册失败: ' + (data.message || '未知错误');
        }
    } catch (error) {
        const errorMessage = error.response?.data?.message || error.message || '网络错误';
        responseDiv.textContent = '错误: ' + errorMessage;
    }
}

// ==================== 用户信息 ====================
function renderAvatar(element, avatarUrl, username) {
    if (avatarUrl) {
        element.innerHTML = `<img src="${avatarUrl}" alt="${username}" onerror="this.style.display='none'; this.nextSibling.style.display='flex';">
                             <span style="display:none;">${username ? username.charAt(0).toUpperCase() : '?'}</span>`;
    } else {
        element.innerHTML = `<span>${username ? username.charAt(0).toUpperCase() : '?'}</span>`;
    }
}

async function getSelfUserInfo() {
    try {
        const { data } = await axios.get('/api/selfuserinfo');

        if (data.code === 0) {
            userInfo = data.data;
            // 缓存用户信息
            userCache.set(userInfo.id, {
                username: userInfo.username,
                avatar: userInfo.avatar
            });
            document.getElementById('profile-name').textContent = userInfo.username;
            renderAvatar(document.getElementById('profile-avatar'), userInfo.avatar, userInfo.username);
            document.getElementById('profile-id').textContent = 'ID: ' + userInfo.id;
            // 更新详细信息
            updateProfileDetails();
            // 获取用户信息后刷新自己的头像显示
            refreshSelfAvatar();
        }
    } catch (error) {
        console.error('获取个人信息错误:', error);
    }
}

// 更新个人信息详情显示
function updateProfileDetails() {
    if (!userInfo) return;

    // 性别
    const sexMap = { 0: '未知', 1: '男', 2: '女' };
    const sexEl = document.getElementById('profile-sex');
    if (sexEl) {
        sexEl.textContent = sexMap[userInfo.sex] || '未知';
    }

    // 生日
    const birthdayEl = document.getElementById('profile-birthday');
    if (birthdayEl) {
        birthdayEl.textContent = userInfo.birthday || '未设置';
    }

    // 注册时间
    const createTimeEl = document.getElementById('profile-create-time');
    if (createTimeEl) {
        if (userInfo.created_at) {
            const date = new Date(userInfo.created_at * 1000);
            createTimeEl.textContent = date.toLocaleDateString('zh-CN');
        } else {
            createTimeEl.textContent = '-';
        }
    }
}

// 更新用户信息（用于保存后刷新）
function updateUserInfo(newUserInfo) {
    if (!newUserInfo) return;

    userInfo = newUserInfo;
    // 更新缓存
    userCache.set(userInfo.id, {
        username: userInfo.username,
        avatar: userInfo.avatar
    });

    // 更新页面显示
    document.getElementById('profile-name').textContent = userInfo.username;
    renderAvatar(document.getElementById('profile-avatar'), userInfo.avatar, userInfo.username);
    document.getElementById('profile-id').textContent = 'ID: ' + userInfo.id;

    // 更新详细信息
    updateProfileDetails();

    // 刷新自己的头像显示
    refreshSelfAvatar();
}

// ==================== 个人信息编辑 ====================

// 显示个人信息编辑面板
function showProfileEdit() {
    if (!userInfo) {
        alert('请先登录');
        return;
    }

    const html = `
        <div class="profile-edit-panel" id="profile-edit-panel">
            <div class="profile-edit-mask" onclick="closeProfileEdit()"></div>
            <div class="profile-edit-content">
                <div class="profile-edit-header">
                    <button class="profile-edit-close" onclick="closeProfileEdit()">×</button>
                    <h3 class="profile-edit-title">编辑个人信息</h3>
                    <button class="profile-edit-save" onclick="saveProfile()">保存</button>
                </div>
                <div class="profile-edit-body">
                    <div class="profile-edit-item">
                        <label>用户名</label>
                        <input type="text" id="edit-username" value="${userInfo.username || ''}" maxlength="20" placeholder="请输入用户名">
                    </div>
                    <div class="profile-edit-item">
                        <label>个人简介</label>
                        <textarea id="edit-intro" maxlength="255" placeholder="介绍一下自己吧~" rows="3">${userInfo.intro || ''}</textarea>
                    </div>
                    <div class="profile-edit-item">
                        <label>性别</label>
                        <select id="edit-sex" class="form-select">
                            <option value="0" ${userInfo.sex === 0 ? 'selected' : ''}>未知</option>
                            <option value="1" ${userInfo.sex === 1 ? 'selected' : ''}>男</option>
                            <option value="2" ${userInfo.sex === 2 ? 'selected' : ''}>女</option>
                        </select>
                    </div>
                    <div class="profile-edit-item">
                        <label>生日</label>
                        <input type="date" id="edit-birthday" class="form-date" value="${userInfo.birthday || ''}">
                    </div>
                    <div class="profile-edit-item">
                        <label>手机号</label>
                        <input type="tel" id="edit-phone" value="${userInfo.phone || ''}" maxlength="16" placeholder="请输入手机号">
                    </div>
                    <div class="profile-edit-item">
                        <label>邮箱</label>
                        <input type="email" id="edit-email" value="${userInfo.email || ''}" maxlength="64" placeholder="请输入邮箱">
                    </div>
                </div>
            </div>
        </div>
    `;

    // 添加到body
    const div = document.createElement('div');
    div.innerHTML = html;
    document.body.appendChild(div);

    // 添加样式
    const style = document.createElement('style');
    style.textContent = `
        .profile-edit-panel { position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 1000; }
        .profile-edit-mask { position: absolute; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); }
        .profile-edit-content { position: absolute; bottom: 0; left: 0; right: 0; background: #fff; border-radius: 16px 16px 0 0; max-height: 85vh; overflow-y: auto; }
        .profile-edit-header { display: flex; align-items: center; justify-content: space-between; padding: 16px; border-bottom: 1px solid #eee; position: sticky; top: 0; background: #fff; }
        .profile-edit-title { font-size: 16px; font-weight: 500; margin: 0; }
        .profile-edit-close { background: none; border: none; font-size: 24px; color: #999; padding: 0; }
        .profile-edit-save { background: var(--primary-color, #4a90d9); color: #fff; border: none; padding: 8px 16px; border-radius: 4px; font-size: 14px; }
        .profile-edit-body { padding: 20px 16px; }
        .profile-edit-item { margin-bottom: 20px; }
        .profile-edit-item label { display: block; font-size: 14px; color: #666; margin-bottom: 8px; }
        .profile-edit-item input, .profile-edit-item select, .profile-edit-item textarea { width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 8px; font-size: 16px; box-sizing: border-box; resize: none; }
        .profile-edit-item input:focus, .profile-edit-item select:focus, .profile-edit-item textarea:focus { outline: none; border-color: var(--primary-color, #4a90d9); }
    `;
    document.head.appendChild(style);
}

// 关闭个人信息编辑面板
function closeProfileEdit() {
    const panel = document.getElementById('profile-edit-panel');
    if (panel) {
        panel.remove();
    }
}

// 保存个人信息
async function saveProfile() {
    const username = document.getElementById('edit-username').value.trim();
    const intro = document.getElementById('edit-intro').value.trim();
    const sex = parseInt(document.getElementById('edit-sex').value);
    const birthday = document.getElementById('edit-birthday').value;
    const phone = document.getElementById('edit-phone').value.trim();
    const email = document.getElementById('edit-email').value.trim();

    if (!username || username.length < 2) {
        alert('用户名至少需要2个字符');
        return;
    }

    // 校验手机号格式（如果填写了）
    if (phone && !/^1[3-9]\d{9}$/.test(phone)) {
        alert('请输入正确的手机号');
        return;
    }

    // 校验邮箱格式（如果填写了）
    if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
        alert('请输入正确的邮箱地址');
        return;
    }

    // 防无变化检测：检查是否有任何变更
    const hasChanged = (
        username !== (userInfo.username || '') ||
        intro !== (userInfo.intro || '') ||
        sex !== userInfo.sex ||
        birthday !== (userInfo.birthday || '') ||
        phone !== (userInfo.phone || '') ||
        email !== (userInfo.email || '')
    );

    if (!hasChanged) {
        closeProfileEdit();
        return;
    }

    try {
        const { data } = await axios.post('/api/updateuserinfo', {
            username: username,
            intro: intro,
            sex: sex,
            birthday: birthday || null,
            phone: phone || null,
            email: email || null
        });

        if (data.code === 0) {
            closeProfileEdit();
            // 如果返回了最新数据，直接使用；否则重新获取
            if (data.data) {
                updateUserInfo(data.data);
            } else {
                await getSelfUserInfo();
            }
        } else {
            alert(data.message || '保存失败');
        }
    } catch (error) {
        console.error('保存个人信息错误:', error);
        alert('保存失败，请重试');
    }
}
// 显示好友信息弹窗
async function showFriendInfo(friendId) {
    try {
        const { data } = await axios.get('/api/otheruserinfo', {
            params: { user_id: friendId }
        });

        if (data.code === 0 && data.data) {
            const info = data.data;
            const sexMap = { 0: '未知', 1: '男', 2: '女' };

            const html = `
                <div class="friend-info-panel" id="friend-info-panel">
                    <div class="friend-info-mask" onclick="closeFriendInfo()"></div>
                    <div class="friend-info-content">
                        <div class="friend-info-header">
                            <button class="friend-info-close" onclick="closeFriendInfo()">×</button>
                            <h3 class="friend-info-title">好友信息</h3>
                            <div style="width: 40px;"></div>
                        </div>
                        <div class="friend-info-body">
                            <div class="friend-info-avatar">
                                ${info.avatar
                                    ? `<img src="${info.avatar}" alt="${info.username}" onerror="this.parentElement.innerHTML='${info.username ? info.username.charAt(0).toUpperCase() : '?'}'">`
                                    : `<span>${info.username ? info.username.charAt(0).toUpperCase() : '?'}</span>`
                                }
                            </div>
                            <div class="friend-info-username">${info.username || '未知用户'}</div>
                            <div class="friend-info-details">
                                <div class="friend-info-item">
                                    <span class="friend-info-label">性别</span>
                                    <span class="friend-info-value">${sexMap[info.sex] || '未知'}</span>
                                </div>
                                <div class="friend-info-item">
                                    <span class="friend-info-label">生日</span>
                                    <span class="friend-info-value">${info.birthday || '未设置'}</span>
                                </div>
                                <div class="friend-info-item">
                                    <span class="friend-info-label">个人简介</span>
                                    <span class="friend-info-value intro">${info.intro || '这个人很懒，什么都没写~'}</span>
                                </div>
                            </div>
                            <div class="friend-info-actions">
                                <button class="friend-info-chat-btn" onclick="closeFriendInfo(); openChat(${info.id}, '${info.username}')">发消息</button>
                            </div>
                        </div>
                    </div>
                </div>
            `;

            // 移除已存在的弹窗
            const existingPanel = document.getElementById('friend-info-panel');
            if (existingPanel) {
                existingPanel.remove();
            }

            // 添加到body
            const div = document.createElement('div');
            div.innerHTML = html;
            document.body.appendChild(div);

            // 添加样式
            if (!document.getElementById('friend-info-style')) {
                const style = document.createElement('style');
                style.id = 'friend-info-style';
                style.textContent = `
                    .friend-info-panel { position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 1000; }
                    .friend-info-mask { position: absolute; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); }
                    .friend-info-content { position: absolute; bottom: 0; left: 0; right: 0; background: #fff; border-radius: 16px 16px 0 0; max-height: 70vh; overflow-y: auto; }
                    .friend-info-header { display: flex; align-items: center; justify-content: space-between; padding: 16px; border-bottom: 1px solid #eee; position: sticky; top: 0; background: #fff; }
                    .friend-info-title { font-size: 16px; font-weight: 500; margin: 0; }
                    .friend-info-close { background: none; border: none; font-size: 24px; color: #999; padding: 0; }
                    .friend-info-body { padding: 20px 16px; text-align: center; }
                    .friend-info-avatar { width: 80px; height: 80px; border-radius: 50%; background: linear-gradient(135deg, #667eea, #764ba2); color: white; font-size: 32px; font-weight: 600; display: flex; align-items: center; justify-content: center; margin: 0 auto 16px; overflow: hidden; }
                    .friend-info-avatar img { width: 100%; height: 100%; object-fit: cover; }
                    .friend-info-username { font-size: 20px; font-weight: 600; margin-bottom: 20px; }
                    .friend-info-details { text-align: left; background: #f8f9fa; border-radius: 8px; padding: 16px; margin-bottom: 20px; }
                    .friend-info-item { display: flex; padding: 10px 0; border-bottom: 1px solid #eee; }
                    .friend-info-item:last-child { border-bottom: none; }
                    .friend-info-label { width: 70px; color: #666; font-size: 14px; }
                    .friend-info-value { flex: 1; color: #333; font-size: 14px; }
                    .friend-info-value.intro { word-break: break-all; }
                    .friend-info-actions { display: flex; gap: 12px; }
                    .friend-info-chat-btn { flex: 1; padding: 12px; background: var(--primary-color, #4a90d9); color: #fff; border: none; border-radius: 8px; font-size: 16px; cursor: pointer; }
                    .friend-info-chat-btn:active { opacity: 0.8; }
                    @media screen and (min-width: 500px) {
                        .friend-info-content { width: 350px; left: 50%; transform: translateX(-50%); border-radius: 16px; }
                    }
                `;
                document.head.appendChild(style);
            }
        } else {
            alert('获取好友信息失败');
        }
    } catch (error) {
        console.error('获取好友信息错误:', error);
        alert('获取好友信息失败，请重试');
    }
}

// 关闭好友信息弹窗
function closeFriendInfo() {
    const panel = document.getElementById('friend-info-panel');
    if (panel) {
        panel.remove();
    }
}

async function uploadAvatar() {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'image/jpeg,image/png,image/gif,image/webp';
    input.onchange = async function (e) {
        const file = e.target.files[0];
        if (!file) return;

        // 检查文件大小 (2MB)
        if (file.size > 2 * 1024 * 1024) {
            alert('图片大小不能超过2MB');
            return;
        }

        const formData = new FormData();
        formData.append('avatar', file);

        try {
            const responseDiv = document.getElementById('profile-response');
            responseDiv.textContent = '上传中...';
            const { data } = await axios.post('/api/uploadavatar', formData, {
                headers: {
                    'Content-Type': 'multipart/form-data'
                }
            });

            if (data.code === 0) {
                alert('头像上传成功！');
                await getSelfUserInfo();
                // 上传成功后刷新自己的头像显示
                refreshSelfAvatar();
            } else {
                responseDiv.textContent = '上传失败: ' + (data.message || '未知错误');
            }
        } catch (error) {
            const errorMessage = error.response?.data?.message || error.message || '网络错误';
            document.getElementById('profile-response').textContent = '错误: ' + errorMessage;
        }
    };
    input.click();
}

// 刷新页面上显示的自己头像
function refreshSelfAvatar() {
    const selfUserId = userInfo?.user_id;
    if (!selfUserId) return;

    const newAvatar = userInfo?.avatar;
    if (!newAvatar) return;

    // 更新个人资料页的头像
    const profileAvatarImg = document.querySelector('#profile-page .avatar-img');
    if (profileAvatarImg) {
        profileAvatarImg.src = newAvatar;
    }

    // 更新聊天消息中自己的头像
    document.querySelectorAll('.message-item.sent .msg-avatar img').forEach(img => {
        const src = img.src;
        const userIdMatch = src.match(/\/avatar\/(\d+)/);
        if (userIdMatch) {
            const userId = parseInt(userIdMatch[1]);
            if (userId === selfUserId) {
                img.src = newAvatar;
            }
        }
    });

    // 更新聊天页面头部的头像
    const chatHeaderAvatar = document.querySelector('.chat-header .avatar-img');
    if (chatHeaderAvatar && chatHeaderAvatar.src.includes(`/avatar/${selfUserId}`)) {
        chatHeaderAvatar.src = newAvatar;
    }

    // 更新底部导航栏的头像
    const tabAvatar = document.querySelector('.tab-item img');
    if (tabAvatar) {
        tabAvatar.src = newAvatar;
    }
}

// ==================== 好友列表 ====================
async function getFriendList() {
    try {
        const { data } = await axios.get('/api/friendlist');

        if (data.code === 0) {
            const friendList = data.data || [];
            // 缓存好友信息
            userCache.setMulti(friendList);
            cache.friendList = friendList;
            renderFriendList(friendList);
        }
    } catch (error) {
        console.error('获取好友列表错误:', error);
    }
}

function renderFriendList(friendList) {
    const friendListElement = document.getElementById('friend-list');
    const isFriendTabVisible = !document.getElementById('friend-tab').classList.contains('hidden');

    if (!isFriendTabVisible) return;

    if (!friendList || friendList.length === 0) {
        friendListElement.innerHTML = `
            <div class="empty-state">
                <div class="empty-state-icon">👥</div>
                <div class="empty-state-text">暂无好友</div>
            </div>
        `;
        return;
    }

    friendListElement.innerHTML = '';
    friendList.forEach(friend => {
        const li = document.createElement('li');
        li.className = 'friend-item';
        li.id = 'friend-item-' + friend.id;

        // 使用缓存的在线状态
        const isOnline = onlineStatusCache.get(friend.id) ?? friend.online;

        const avatarHtml = friend.avatar
            ? `<img src="${friend.avatar}" alt="${friend.username}" onerror="this.parentElement.innerHTML='${friend.username ? friend.username.charAt(0).toUpperCase() : '?'}'">`
            : `<span>${friend.username ? friend.username.charAt(0).toUpperCase() : '?'}</span>`;

        li.innerHTML = `
            <div class="friend-avatar" onclick="event.stopPropagation(); showFriendInfo(${friend.id})">${avatarHtml}</div>
            <div class="friend-info" onclick="openChat(${friend.id}, '${friend.username}')">
                <div class="friend-name">${friend.username || '未知用户'}</div>
                <div class="friend-status ${isOnline ? 'online' : ''}">${isOnline ? '🟢 在线' : '⚫ 离线'}</div>
            </div>
            <button class="friend-delete-btn" onclick="event.stopPropagation(); deleteFriend(${friend.id})" title="删除好友">🗑</button>
        `;
        friendListElement.appendChild(li);
    });
}

// ==================== 好友请求列表（好友Tab显示） ====================

// 获取好友请求列表并渲染到好友Tab
async function getFriendRequestListForUI() {
    try {
        const { data } = await axios.get('/api/friendreqlist');
        const requestList = data.data || [];

        // 更新未读数
        if (requestList.length > 0) {
            unreadFriendRequestCount = requestList.length;
            updateFriendRequestBadge();
        }

        // 更新好友请求数量显示
        const countEl = document.getElementById('friend-request-count');
        if (countEl) {
            countEl.textContent = requestList.length;
            if (requestList.length > 0) {
                countEl.classList.add('show');
            } else {
                countEl.classList.remove('show');
            }
        }

        // 获取并缓存请求者信息
        if (requestList.length > 0) {
            await userCache.fetchMissing(requestList);
        }

        renderFriendRequestList(requestList);
    } catch (error) {
        console.error('获取好友请求列表错误:', error);
    }
}

// 渲染好友请求列表（好友Tab）
function renderFriendRequestList(requestList) {
    const listEl = document.getElementById('friend-request-list');
    if (!listEl) return;

    if (requestList.length === 0) {
        listEl.innerHTML = `
            <div class="empty-state small">
                <div class="empty-state-icon">📭</div>
                <div class="empty-state-text">暂无好友请求</div>
            </div>
        `;
        return;
    }

    listEl.innerHTML = '';
    requestList.forEach(userId => {
        const user = userCache.get(userId);
        const username = user ? user.username : `用户${userId}`;
        const avatar = user ? user.avatar : '';
        const initial = username ? username.charAt(0).toUpperCase() : '?';
        const avatarHtml = avatar
            ? `<img src="${avatar}" alt="${initial}" onerror="this.style.display='none'; this.nextSibling.style.display='flex';">
               <span style="display:none;">${initial}</span>`
            : `<span>${initial}</span>`;

        const li = document.createElement('li');
        li.className = 'friend-request-item';
        li.innerHTML = `
            <div class="friend-request-avatar">${avatarHtml}</div>
            <div class="friend-request-info">
                <div class="friend-request-name">${username}</div>
                <div class="friend-request-id">ID: ${userId}</div>
            </div>
            <div class="friend-request-actions">
                <button class="accept-btn" onclick="acceptFriendRequest(${userId})">接受</button>
                <button class="reject-btn" onclick="rejectFriendRequest(${userId})">拒绝</button>
            </div>
        `;
        listEl.appendChild(li);
    });
}
async function sendFriendRequest() {
    const friendIDInput = document.getElementById('friend-id').value;
    const friendID = parseInt(friendIDInput);
    const responseDiv = document.getElementById('add-friend-response');

    if (isNaN(friendID) || friendID === 0) {
        responseDiv.textContent = '请输入有效的用户ID';
        return;
    }

    try {
        responseDiv.textContent = '发送中...';
        const { data } = await axios.post('/api/newfriendreq', { friend_id: friendID });

        if (data.code === 0) {
            alert('好友请求发送成功！');
        } else {
            responseDiv.textContent = '发送好友请求失败: ' + (data.message || '未知错误');
        }
    } catch (error) {
        const errorMessage = error.response?.data?.message || error.message || '网络错误';
        responseDiv.textContent = '错误: ' + errorMessage;
    }
}

async function getFriendRequestList() {
    try {
        const { data } = await axios.get('/api/friendreqlist');

        if (data.code === 0) {
            const requestList = data.data || [];
            if (requestList.length > 0) {
                unreadFriendRequestCount = requestList.length;
                updateFriendRequestBadge();
                // 获取并缓存请求者信息
                await userCache.fetchMissing(requestList);
            }
            return requestList;
        }
        return [];
    } catch (error) {
        console.error('获取好友请求列表错误:', error);
        return [];
    }
}

async function acceptFriendRequest(fromID) {
    try {
        const { data } = await axios.post('/api/okwithfriendreq', { from_id: fromID });

        if (data.code === 0) {
            alert('已接受好友请求！');
            getFriendRequestList();
            getFriendRequestListForUI();
        } else {
            alert('接受好友请求失败: ' + (data.message || '未知错误'));
        }
    } catch (error) {
        console.error('接受好友请求错误:', error);
    }
}

// 删除好友
async function deleteFriend(friendId) {
    if (!confirm('确定要删除该好友吗？')) return;

    try {
        const { data } = await axios.post('/api/deletefriend', { friend_id: friendId });

        if (data.code === 0) {
            alert('已删除好友');
            getFriendList();
            // 从最近联系人中移除
            removeFromRecentChats(friendId);
            renderMessageList(getLocalRecentChats());
        } else {
            alert('删除好友失败: ' + (data.message || '未知错误'));
        }
    } catch (error) {
        console.error('删除好友错误:', error);
    }
}

// 从最近联系人中移除
function removeFromRecentChats(friendId) {
    let recentChats = getLocalRecentChats();
    recentChats = recentChats.filter(c => c.friend_id !== friendId);
    saveLocalRecentChats(recentChats);
}

// 拒绝好友请求
async function rejectFriendRequest(fromID) {
    try {
        const { data } = await axios.post('/api/rejectfriendreq', { from_id: fromID });

        if (data.code === 0) {
            alert('已拒绝好友请求');
            getFriendRequestList();
            getFriendRequestListForUI();
        } else {
            alert('拒绝好友请求失败: ' + (data.message || '未知错误'));
        }
    } catch (error) {
        console.error('拒绝好友请求错误:', error);
    }
}

// ==================== 页面关闭/刷新处理 ====================
// 页面关闭时发送同步下线请求（keepalive 确保持久连接）
window.addEventListener('beforeunload', async (e) => {
    if (!getToken()) return;
    
    // 发送下线请求，使用 keepalive 确保请求被发送
    try {
        await fetch('/api/offline', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + getToken(),
                'Content-Type': 'application/json'
            },
            keepalive: true,
            signal: AbortSignal.timeout(3000)
        });
    } catch (e) {
        // 忽略错误，后端 WS 断开时会自动清理
    }
});

function logout() {
    // 先发送下线请求
    const token = getToken();
    if (token) {
        axios.post('/api/offline').catch(() => {});
    }
    
    // 清除本地缓存
    clearUserLocalCache();
    
    deleteCookie('token');
    delete axios.defaults.headers.common['Authorization'];
    document.getElementById('login-page').classList.remove('hidden');
    document.getElementById('main-page').classList.add('hidden');
    if (ws) {
        ws.close();
        ws = null;
    }
    stopHeartbeat();
    if (friendStatusTimer) {
        clearInterval(friendStatusTimer);
    }
    alert('已退出登录');
}

// ==================== 消息列表 ====================
async function getMessageList() {
    // 从本地存储获取最近联系人
    const localChats = getLocalRecentChats();

    if (localChats.length > 0) {
        // 缓存用户信息
        localChats.forEach(msg => {
            userCache.set(msg.friend_id, {
                username: msg.friend_name,
                avatar: msg.friend_avatar
            });
        });
        cache.messageList = localChats;
        renderMessageList(localChats);
    } else {
        // 没有本地数据时显示空状态
        renderMessageList([]);
    }
}

function renderMessageList(messageList) {
    const messageListElement = document.getElementById('message-list');
    const isMessageTabVisible = !document.getElementById('message-tab').classList.contains('hidden');

    if (!isMessageTabVisible) return;

    messageListElement.innerHTML = '';

    if (!messageList || messageList.length === 0) {
        const noMsgDiv = document.createElement('div');
        noMsgDiv.className = 'empty-state';
        noMsgDiv.innerHTML = `
            <div class="empty-state-icon">💬</div>
            <div class="empty-state-text">暂无消息，点击好友开始聊天</div>
        `;
        messageListElement.appendChild(noMsgDiv);
        return;
    }

    messageList.forEach(msg => {
        const li = document.createElement('li');
        li.className = 'message-item';
        li.id = 'message-item-' + msg.friend_id;
        li.onclick = () => openChat(msg.friend_id, msg.friend_name);

        // 使用缓存的在线状态
        const isOnline = onlineStatusCache.get(msg.friend_id) ?? msg.online;

        const avatarHtml = msg.friend_avatar
            ? `<img src="${msg.friend_avatar}" alt="${msg.friend_name}" onerror="this.parentElement.innerHTML='${msg.friend_name ? msg.friend_name.charAt(0).toUpperCase() : '?'}'">`
            : `<span>${msg.friend_name ? msg.friend_name.charAt(0).toUpperCase() : '?'}</span>`;

        li.innerHTML = `
            <div class="message-avatar">${avatarHtml}</div>
            <div class="message-content">
                <div class="message-header">
                    <span class="message-name">${msg.friend_name || '未知用户'}</span>
                    <span class="message-time ${isOnline ? 'online' : ''}">${isOnline ? '🟢' : '⚫'}</span>
                </div>
                <div class="message-preview">${msg.last_message || ''}</div>
            </div>
        `;
        messageListElement.appendChild(li);
    });
}

// ==================== 聊天功能 ====================
async function loadChatHistory(userId) {
    const chatContent = document.getElementById('chat-content');

    // 优先从本地存储加载（即时响应）
    const localHistory = getLocalChatHistory(userId);
    if (localHistory && localHistory.length > 0) {
        chatContent.innerHTML = '';
        localHistory.forEach(msg => {
            appendMessage(msg, null);
        });
        // 滚动到底部，等待图片加载完成后再滚动一次
        scrollToBottom();
    } else {
        chatContent.innerHTML = '<div class="empty-state"><div class="empty-state-icon">💭</div><div class="empty-state-text">暂无聊天记录，开始聊天吧</div></div>';
    }

    // 不再从服务端加载历史记录（改为WebSocket同步）
    // 服务端只存储离线消息，用户上线时会通过WebSocket收到
}

// 滚动到聊天底部
function scrollToBottom() {
    requestAnimationFrame(() => {
        const chatContent = document.getElementById('chat-content');
        if (chatContent) {
            // 立即滚动一次
            chatContent.scrollTop = chatContent.scrollHeight;
            // 等待图片加载完成后再滚动
            const images = chatContent.querySelectorAll('img');
            if (images.length > 0) {
                let loadedCount = 0;
                images.forEach(img => {
                    if (img.complete) {
                        loadedCount++;
                    } else {
                        img.onload = () => {
                            loadedCount++;
                            if (loadedCount === images.length) {
                                chatContent.scrollTop = chatContent.scrollHeight;
                            }
                        };
                        img.onerror = () => {
                            loadedCount++;
                            if (loadedCount === images.length) {
                                chatContent.scrollTop = chatContent.scrollHeight;
                            }
                        };
                    }
                });
                // 如果所有图片都已加载
                if (loadedCount === images.length) {
                    chatContent.scrollTop = chatContent.scrollHeight;
                }
            }
        }
    });
}

function appendMessage(msg, isOwn = null) {
    const chatContent = document.getElementById('chat-content');
    if (isOwn === null) {
        isOwn = msg.from_id === currentUserId;
    }

    // 去重：如果消息已存在则不添加
    if (msg.msg_id) {
        const existing = document.querySelector(`[data-msg-id="${msg.msg_id}"]`);
        if (existing) {
            console.log(`消息 ${msg.msg_id} 已存在，跳过`);
            return;
        }
    }

    // 从缓存获取用户信息
    const userInfo = userCache.get(msg.from_id);
    const username = userInfo ? userInfo.username : `用户${msg.from_id}`;
    const avatar = userInfo ? userInfo.avatar : '';

    // 生成头像HTML
    const avatarHtml = generateAvatarHtml(avatar, username, isOwn);

    const messageDiv = document.createElement('div');
    messageDiv.className = 'message ' + (isOwn ? 'own' : 'other');

    // 设置消息ID用于去重
    if (msg.msg_id) {
        messageDiv.setAttribute('data-msg-id', msg.msg_id);
    }

    // 消息状态指示器（仅发送者显示，仅失败时显示）
    let statusHtml = '';
    if (isOwn && msg.msg_id && msg.send_status === 'failed') {
        statusHtml = '<span class="msg-status failed">✗</span>';
    }

    // 处理文件消息（message_type: 2=文件, 3=图片）
    let contentHtml = escapeHtml(msg.content || '');
    if ((msg.message_type === 2 || msg.message_type === 3) && msg.file_url) {
        const fileName = msg.file_name || '文件';
        const ext = fileName.split('.').pop().toLowerCase();
        const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'];
        const isImage = imageExts.includes(ext) || msg.message_type === 3;

        // 构建API下载链接
        const buildFileUrl = () => {
            const fileUrl = msg.file_url;
            // 如果已经是完整URL，直接返回
            if (fileUrl.startsWith('http://') || fileUrl.startsWith('https://') || fileUrl.startsWith('/')) {
                return fileUrl;
            }
            // 根据消息类型确定file_type
            const fileType = msg.message_type === 3 ? 'image' : 'file';
            // 否则构建API路径
            return `/api/downloadfile?file_name=${encodeURIComponent(fileUrl)}&file_type=${fileType}&from_id=${msg.from_id}`;
        };

        if (isImage) {
            // 图片消息 - 可预览
            const imageUrl = buildFileUrl();
            const escapedUrl = imageUrl.replace(/'/g, "\\'");
            contentHtml = `<img src="${imageUrl}" class="file-preview" alt="${escapeHtml(fileName)}" data-full-url="${escapedUrl}" onclick="openImagePreview(this.getAttribute('data-full-url'))">`;
        } else {
            // 其他文件 - 下载
            const downloadUrl = buildFileUrl();
            contentHtml = `<div class="file-message">
                <span class="file-icon">📄</span>
                <span class="file-name">${escapeHtml(fileName)}</span>
                <a href="${downloadUrl}" download="${escapeHtml(fileName)}" class="file-download" onclick="event.stopPropagation()">下载</a>
            </div>`;
        }
    }

    messageDiv.innerHTML = `
        <div class="message-avatar-small">${avatarHtml}</div>
        <div class="message-bubble">${contentHtml}</div>
        ${statusHtml}
    `;
    chatContent.appendChild(messageDiv);
    // 滚动到底部，等待图片加载
    scrollToBottom();
}

// 生成头像HTML
function generateAvatarHtml(avatar, username, isOwn = false) {
    const initial = username ? username.charAt(0).toUpperCase() : '?';
    const alt = isOwn ? '我' : initial;

    if (avatar) {
        return `<img src="${avatar}" alt="${alt}" onerror="this.style.display='none'; this.nextSibling.style.display='flex';">
                <span style="display:none;">${alt}</span>`;
    }
    return `<span>${alt}</span>`;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 打开图片预览
function openImagePreview(url) {
    const overlay = document.getElementById('image-preview-overlay');
    const img = document.getElementById('preview-image');
    if (overlay && img) {
        img.src = url;
        overlay.classList.remove('hidden');
    }
}

// 关闭图片预览
function closeImagePreview() {
    const overlay = document.getElementById('image-preview-overlay');
    if (overlay) {
        overlay.classList.add('hidden');
    }
}

// 生成唯一消息ID
function generateMsgId() {
    return 'msg_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
}

function sendChatMessage() {
    const messageInput = document.getElementById('chat-message');
    const text = messageInput.value.trim();
    if (!text) return;

    const sendTime = Date.now();
    const msgId = generateMsgId();
    
    // 保存消息到本地历史
    const msgData = {
        msg_id: msgId,
        from_id: currentUserId, 
        content: text, 
        send_time: sendTime,
        message_type: 1
    };

    // 先显示到界面
    appendMessage(msgData, true);
    messageInput.value = '';

    // 本地存储消息
    addMessageToLocalHistory(currentChatId, {
        ...msgData,
        to_id: currentChatId
    });

    // 更新最近联系人
    const friendName = userCache.getUsername(currentChatId) || `用户${currentChatId}`;
    const friendAvatar = userCache.getAvatar(currentChatId) || '';
    updateLocalRecentChat(currentChatId, friendName, friendAvatar, text, sendTime);

    // 发送WebSocket消息
    if (ws && ws.readyState === WebSocket.OPEN) {
        const wsData = {
            msg_id: msgId,
            from_id: currentUserId,
            to_id: currentChatId,
            room_id: 0,
            message_type: 1,
            content: text,
            file_id: 0
        };
        ws.send(JSON.stringify(wsData));
    } else {
        connectWebSocket();
    }
}

// 处理文件选择
async function handleFileSelect(file) {
    if (!file) return;

    const isImage = file.type.startsWith('image/');
    const MAX_IMAGE_SIZE = 2 * 1024 * 1024;  // 2MB
    const MAX_IMAGE_DIM = 1920;  // 最大边长1920px

    // 如果是图片，压缩处理
    if (isImage) {
        file = await compressImage(file, MAX_IMAGE_DIM, MAX_IMAGE_SIZE);
    }

    // 检查文件大小（40MB限制）
    if (file.size > 40 * 1024 * 1024) {
        alert('文件大小不能超过40MB');
        return;
    }

    const confirmed = confirm(`确定要发送文件 "${file.name}" (${formatFileSize(file.size)}) 吗？`);
    if (!confirmed) {
        document.getElementById('file-input').value = '';
        return;
    }

    try {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('type', isImage ? 'image' : 'file');  // 文件类型

        const responseDiv = document.getElementById('login-response');
        responseDiv.textContent = '上传文件中...';

        const { data } = await axios.post('/api/uploadfile', formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        });

        if (data.code === 0) {
            const fileName = data.data;
            if (isImage) {
                sendImageMessage(fileName, file.name, file.size);
            } else {
                sendFileMessage(fileName, file.name, file.size);
            }
        } else {
            alert('文件上传失败: ' + (data.message || '未知错误'));
        }
    } catch (error) {
        console.error('文件上传错误:', error);
        alert('文件上传失败');
    } finally {
        document.getElementById('file-input').value = '';
    }
}

// 图片压缩函数
function compressImage(file, maxDim, maxSize) {
    return new Promise((resolve) => {
        const img = new Image();
        img.onload = () => {
            // 计算压缩后尺寸
            let width = img.width;
            let height = img.height;
            if (width > maxDim || height > maxDim) {
                if (width > height) {
                    height = Math.round(height * maxDim / width);
                    width = maxDim;
                } else {
                    width = Math.round(width * maxDim / height);
                    height = maxDim;
                }
            }

            // 如果文件已经小于限制，直接返回
            if (file.size <= maxSize) {
                URL.revokeObjectURL(img.src);
                resolve(file);
                return;
            }

            // 压缩到指定大小
            const canvas = document.createElement('canvas');
            canvas.width = width;
            canvas.height = height;
            const ctx = canvas.getContext('2d');
            ctx.drawImage(img, 0, 0, width, height);

            // 逐步降低质量直到满足大小限制
            let quality = 0.9;
            const step = 0.1;

            function tryCompress() {
                canvas.toBlob((blob) => {
                    if (blob.size <= maxSize || quality <= 0.1) {
                        const compressedFile = new File([blob], file.name, { type: file.type });
                        URL.revokeObjectURL(img.src);
                        resolve(compressedFile);
                    } else {
                        quality -= step;
                        tryCompress();
                    }
                }, file.type, quality);
            }
            tryCompress();
        };
        img.src = URL.createObjectURL(file);
    });
}

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

// 发送图片消息
async function sendImageMessage(fileUrl, fileName, fileSize) {
    const msgId = generateMsgId();
    const sendTime = Date.now();

    // 构建图片消息数据
    const msgData = {
        msg_id: msgId,
        from_id: currentUserId,
        to_id: currentChatId,
        content: fileUrl,
        file_url: fileUrl,
        file_name: fileName,
        file_size: fileSize,
        send_time: sendTime,
        message_type: 3  // 图片类型
    };

    // 显示图片消息
    appendMessage(msgData, true);

    // 保存到本地历史
    addMessageToLocalHistory(currentChatId, msgData);

    // 更新最近联系人
    const friendName = userCache.getUsername(currentChatId) || `用户${currentChatId}`;
    const friendAvatar = userCache.getAvatar(currentChatId) || '';
    updateLocalRecentChat(currentChatId, friendName, friendAvatar, '[图片]', sendTime);

    // 发送WebSocket消息
    if (ws && ws.readyState === WebSocket.OPEN) {
        const wsData = {
            msg_id: msgId,
            from_id: currentUserId,
            to_id: currentChatId,
            room_id: 0,
            message_type: 3,
            content: fileUrl,
            file_url: fileUrl,
            file_name: fileName,
            file_id: 0
        };
        ws.send(JSON.stringify(wsData));
    } else {
        connectWebSocket();
    }
}

// 发送文件消息
async function sendFileMessage(fileUrl, fileName, fileSize) {
    const msgId = generateMsgId();
    const sendTime = Date.now();

    // 构建文件消息数据
    const msgData = {
        msg_id: msgId,
        from_id: currentUserId,
        to_id: currentChatId,
        content: `[文件] ${fileName}`,
        file_url: fileUrl,
        file_name: fileName,
        file_size: fileSize,
        send_time: sendTime,
        message_type: 2
    };

    // 显示文件消息
    appendMessage(msgData, true);

    // 保存到本地历史
    addMessageToLocalHistory(currentChatId, msgData);

    // 更新最近联系人
    const friendName = userCache.getUsername(currentChatId) || `用户${currentChatId}`;
    const friendAvatar = userCache.getAvatar(currentChatId) || '';
    updateLocalRecentChat(currentChatId, friendName, friendAvatar, `[文件] ${fileName}`, sendTime);

    // 发送WebSocket消息
    if (ws && ws.readyState === WebSocket.OPEN) {
        const wsData = {
            msg_id: msgId,
            from_id: currentUserId,
            to_id: currentChatId,
            room_id: 0,
            message_type: 2,
            content: `[文件] ${fileName}`,
            file_url: fileUrl,
            file_name: fileName,
            file_id: 0
        };
        ws.send(JSON.stringify(wsData));
    } else {
        connectWebSocket();
    }
}

// 回车发送消息
document.addEventListener('DOMContentLoaded', function () {
    const chatInput = document.getElementById('chat-message');
    if (chatInput) {
        chatInput.addEventListener('keypress', function (e) {
            if (e.key === 'Enter') {
                e.preventDefault();
                sendChatMessage();
            }
        });
    }
});

// ==================== 好友状态刷新 ====================
function startFriendStatusRefresh() {
    if (friendStatusTimer) {
        clearInterval(friendStatusTimer);
    }

    // 初始加载好友列表时获取在线状态
    if (cache.friendList && cache.friendList.length > 0) {
        requestOnlineStatus();
    }

    // 降低刷新频率，从5秒改为15秒
    friendStatusTimer = setInterval(() => {
        if (getToken()) {
            // 只在需要时刷新好友列表（WebSocket应该处理在线状态更新）
            getFriendList();
        } else {
            clearInterval(friendStatusTimer);
        }
    }, 15000);

    // 启动定时刷新好友头像（5分钟）
    startFriendAvatarRefresh();
}

// 定时刷新好友头像
let friendAvatarTimer = null;
function startFriendAvatarRefresh() {
    if (friendAvatarTimer) {
        clearInterval(friendAvatarTimer);
    }

    // 每5分钟刷新一次好友头像
    friendAvatarTimer = setInterval(async () => {
        if (!getToken()) {
            clearInterval(friendAvatarTimer);
            return;
        }

        try {
            const { data } = await axios.get('/api/friendlist');
            const friendList = data.data || [];
            
            if (friendList && friendList.length > 0) {
                // 更新用户缓存中的头像信息
                friendList.forEach(friend => {
                    userCache.set(friend.user_id, friend);
                });

                // 刷新最近联系人的头像显示
                refreshDisplayedAvatars(friendList);
            }
        } catch (error) {
            console.error('刷新好友头像失败:', error);
        }
    }, 5 * 60 * 1000); // 5分钟
}

// 刷新页面上显示的头像
function refreshDisplayedAvatars(friendList) {
    // 构建 userId -> avatar 的映射
    const avatarMap = new Map();
    friendList.forEach(friend => {
        avatarMap.set(friend.user_id, friend.avatar);
    });

    // 更新好友列表中的头像
    document.querySelectorAll('.friend-item').forEach(item => {
        const userId = parseInt(item.dataset.userId);
        const newAvatar = avatarMap.get(userId);
        if (newAvatar) {
            const avatarImg = item.querySelector('.avatar-img');
            if (avatarImg) {
                avatarImg.src = newAvatar;
            }
        }
    });

    // 更新最近聊天列表中的头像
    document.querySelectorAll('.recent-chat-item').forEach(item => {
        const userId = parseInt(item.dataset.userId);
        const newAvatar = avatarMap.get(userId);
        if (newAvatar) {
            const avatarImg = item.querySelector('.avatar-img');
            if (avatarImg) {
                avatarImg.src = newAvatar;
            }
        }
    });

    // 更新聊天消息中的头像
    document.querySelectorAll('.msg-avatar img').forEach(img => {
        const src = img.src;
        const userIdMatch = src.match(/\/avatar\/(\d+)/);
        if (userIdMatch) {
            const userId = parseInt(userIdMatch[1]);
            const newAvatar = avatarMap.get(userId);
            if (newAvatar && newAvatar !== src) {
                img.src = newAvatar;
            }
        }
    });
}

// 请求在线状态（通过WebSocket）
function requestOnlineStatus() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'get_online_status' }));
    }
}

// ==================== WebSocket ====================
function connectWebSocket() {
    const token = getToken();

    if (!token) {
        console.error('WebSocket连接失败：未登录');
        return;
    }

    if (ws && ws.readyState === WebSocket.OPEN) {
        return;
    }

    const wsUrl = getWsUrl() + '?token=' + token;

    try {
        // 防止重复连接
        if (wsConnecting) {
            console.log('WebSocket正在连接中，跳过');
            return;
        }
        wsConnecting = true;
        ws = new WebSocket(wsUrl);
    } catch (e) {
        console.error('创建WebSocket失败:', e);
        wsConnecting = false;
        return;
    }

    ws.onopen = function () {
        console.log('WebSocket连接成功');
        wsReconnectAttempts = 0;
        wsConnecting = false;
        lastPongTime = Date.now();
        startHeartbeat();

        if ('Notification' in window && Notification.permission === 'default') {
            Notification.requestPermission();
        }
    };

    ws.onerror = function (error) {
        console.error('WebSocket错误:', error);
    };

    ws.onclose = function (event) {
        console.log('WebSocket连接关闭');
        stopHeartbeat();
        ws = null;
        wsConnecting = false;

        if (wsReconnectAttempts < MAX_WS_RECONNECT && getToken()) {
            wsReconnectAttempts++;
            // 添加延迟，避免频繁重连导致服务端的 Add 方法重复关闭连接
            const delay = Math.min(wsReconnectAttempts * 2000, 10000);
            console.log(`${delay/1000}秒后尝试重连...`);
            wsReconnectTimer = setTimeout(connectWebSocket, delay);
        } else if (wsReconnectAttempts >= MAX_WS_RECONNECT) {
            console.warn('WebSocket重连次数已达上限，请刷新页面或重新登录');
        }
    };

    ws.onmessage = function (event) {
        console.log('收到消息:', event.data);
        handleWebSocketMessage(event.data);
    };
}

// ==================== 心跳机制 ====================
function startHeartbeat() {
    stopHeartbeat();
    lastPingTime = Date.now();
    lastPongTime = Date.now();

    heartbeatTimer = setInterval(() => {
        if (ws && ws.readyState === WebSocket.OPEN) {
            // 检查是否超时（基于最后收到pong的时间）
            if (Date.now() - lastPongTime > HEARTBEAT_TIMEOUT) {
                console.warn('心跳超时，连接可能已断开');
                ws.close();
                return;
            }

            // 发送心跳
            try {
                lastPingTime = Date.now();
                ws.send(JSON.stringify({ type: 'ping' }));
                console.log('发送心跳Ping');
            } catch (e) {
                console.error('发送心跳失败:', e);
            }
        }
    }, HEARTBEAT_INTERVAL);
}

function stopHeartbeat() {
    if (heartbeatTimer) {
        clearInterval(heartbeatTimer);
        heartbeatTimer = null;
    }
}

function handleWebSocketMessage(data) {
    try {
        const msg = JSON.parse(data);

        // 处理心跳响应
        if (msg.type === 'pong' || msg.type === 'ping') {
            lastPongTime = Date.now();
            return;
        }

        // 处理消息发送失败
        if (msg.msg_type === 'send_failed') {
            const { msg_id, reason } = msg;
            console.warn('消息发送失败:', msg_id, reason);
            // 更新消息状态为失败
            const msgElement = document.querySelector(`[data-msg-id="${msg_id}"]`);
            if (msgElement) {
                const existingStatus = msgElement.querySelector('.msg-status');
                if (existingStatus) {
                    existingStatus.remove();
                }
                const statusHtml = '<span class="msg-status failed">✗</span>';
                msgElement.insertAdjacentHTML('beforeend', statusHtml);
            }
            return;
        }

        // 处理在线状态更新
        if (msg.type === 'online_status') {
            const { user_id, online } = msg;
            onlineStatusCache.set(user_id, online);
            updateOnlineStatusInUI(user_id, online);
            return;
        }

        // 处理批量在线状态
        if (msg.type === 'batch_online_status') {
            const { statuses } = msg;
            statuses.forEach(({ user_id, online }) => {
                onlineStatusCache.set(user_id, online);
            });
            // 批量更新UI
            if (cache.friendList) {
                renderFriendList(cache.friendList);
            }
            if (cache.messageList) {
                renderMessageList(cache.messageList);
            }
            return;
        }

        if (msg.msg_type === 'friend') {
            // 好友请求通知
            const friendData = msg.data;
            const fromId = friendData.from_id;

            // 如果缓存中没有该用户信息，异步获取
            if (!userCache.get(fromId)) {
                userCache.fetchMissing([fromId]).then(() => {
                    const username = userCache.getUsername(fromId);
                    showFriendRequestToast(username);
                });
            } else {
                const username = userCache.getUsername(fromId);
                showFriendRequestToast(username);
            }

            window._lastFriendRequestFromId = fromId;
            getFriendRequestList();
        } else if (msg.msg_type === 'talk') {
            const msgData = msg.data;
            const fromId = msgData.from_id;
            let content = msgData.content || '';
            const sendTime = msgData.send_time || Date.now();
            const msgId = msgData.msg_id;

            // 本地消息去重：如果已确认消息中已存在，跳过
            if (msgId && confirmedMessages.has(msgId)) {
                console.log(`收到重复消息 ${msgId}，跳过`);
                return;
            }

            // 处理文件消息预览内容
            if (msgData.message_type === 2 || msgData.message_type === 3) {
                const fileName = msgData.file_name || '文件';
                const ext = fileName.split('.').pop().toLowerCase();
                const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'];
                content = imageExts.includes(ext) ? '[图片]' : '[文件] ' + fileName;
            }

            // 标记消息已处理，防止重复显示
            if (msgId) {
                confirmedMessages.add(msgId);
            }

            // 如果缓存中没有该用户信息，异步获取
            if (!userCache.get(fromId)) {
                userCache.fetchMissing([fromId]);
            }

            // 保存到本地历史
            addMessageToLocalHistory(fromId, msgData);

            // 更新最近联系人
            const friendName = userCache.getUsername(fromId) || `用户${fromId}`;
            const friendAvatar = userCache.getAvatar(fromId) || '';
            updateLocalRecentChat(fromId, friendName, friendAvatar, content, sendTime);

            const chatPage = document.getElementById('chat-page');
            const isInChat = !chatPage.classList.contains('hidden') && fromId === currentChatId;

            if (isInChat) {
                // 在聊天界面中显示消息
                appendMessage(msgData, null);
            } else {
                // 不在聊天界面，增加未读计数
                if (!unreadMessages[fromId]) {
                    unreadMessages[fromId] = 0;
                }
                unreadMessages[fromId]++;
                updateUnreadInMessageList(fromId);

                // 显示toast通知
                showMessageToast(fromId, content);

                // 尝试发送浏览器通知
                if (Notification.permission === 'granted') {
                    const notifyUsername = userCache.getUsername(fromId);
                    const notifyAvatar = userCache.getAvatar(fromId);
                    new Notification(notifyUsername + '发来消息', {
                        body: content,
                        icon: notifyAvatar || '💬'
                    });
                }
            }

            // 更新消息列表
            renderMessageList(getLocalRecentChats());

            // 更新总未读消息数量
            unreadMessageCount = Object.values(unreadMessages).reduce((a, b) => a + b, 0);
            updateMessageBadge();
        }
    } catch (error) {
        console.error('解析消息错误:', error);
    }
}

// 更新消息列表中指定联系人的未读数
function updateUnreadInMessageList(friendId) {
    const messageItem = document.getElementById('message-item-' + friendId);
    if (!messageItem) return;

    const unreadCount = unreadMessages[friendId] || 0;
    let badgeEl = messageItem.querySelector('.unread-badge');

    if (unreadCount > 0) {
        if (!badgeEl) {
            badgeEl = document.createElement('span');
            badgeEl.className = 'unread-badge';
            messageItem.appendChild(badgeEl);
        }
        badgeEl.textContent = unreadCount > 99 ? '99+' : unreadCount;
        badgeEl.classList.remove('hidden');
    } else {
        if (badgeEl) {
            badgeEl.classList.add('hidden');
        }
    }
}

// 更新UI中的在线状态
function updateOnlineStatusInUI(userId, isOnline) {
    const friendItem = document.getElementById('friend-item-' + userId);
    const messageItem = document.getElementById('message-item-' + userId);

    const statusHtml = isOnline ? '🟢 在线' : '⚫ 离线';
    const statusClass = isOnline ? 'online' : '';

    if (friendItem) {
        const statusEl = friendItem.querySelector('.friend-status');
        if (statusEl) {
            statusEl.className = 'friend-status ' + statusClass;
            statusEl.textContent = statusHtml;
        }
    }

    if (messageItem) {
        const statusEl = messageItem.querySelector('.message-time');
        if (statusEl) {
            statusEl.className = 'message-time ' + statusClass;
            statusEl.textContent = isOnline ? '🟢' : '⚫';
        }
    }
}

// 自动连接WebSocket
if (getToken()) {
    connectWebSocket();
}

// ==================== 聊天背景设置 ====================

// 背景数据
const BG_SOLID_COLORS = [
    { id: 'solid-1', name: '默认灰', color: '#f5f7fa' },
    { id: 'solid-2', name: '天空蓝', color: '#e3f2fd' },
    { id: 'solid-3', name: '抹茶绿', color: '#e8f5e9' },
    { id: 'solid-4', name: '樱花粉', color: '#fce4ec' },
    { id: 'solid-5', name: '薰衣草', color: '#f3e5f5' },
    { id: 'solid-6', name: '暖阳橙', color: '#fff3e0' },
    { id: 'solid-7', name: '薄荷绿', color: '#e0f2f1' },
    { id: 'solid-8', name: '蜜桃粉', color: '#fde8d8' },
    { id: 'solid-9', name: '淡紫蓝', color: '#e8eaf6' },
    { id: 'solid-10', name: '奶油白', color: '#fefefe' },
];

const BG_GRADIENT_COLORS = [
    { id: 'grad-1', name: '极光', gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' },
    { id: 'grad-2', name: '日出', gradient: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' },
    { id: 'grad-3', name: '森林', gradient: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)' },
    { id: 'grad-4', name: '海洋', gradient: 'linear-gradient(135deg, #2193b0 0%, #6dd5ed 100%)' },
    { id: 'grad-5', name: '落日', gradient: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)' },
    { id: 'grad-6', name: '夜空', gradient: 'linear-gradient(135deg, #0f2027 0%, #203a43 50%, #2c5364 100%)' },
    { id: 'grad-7', name: '梦幻', gradient: 'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)' },
    { id: 'grad-8', name: '烈焰', gradient: 'linear-gradient(135deg, #ff4b1f 0%, #ff9068 100%)' },
    { id: 'grad-9', name: '霓虹', gradient: 'linear-gradient(135deg, #ff0844 0%, #ffb199 100%)' },
    { id: 'grad-10', name: '暖阳', gradient: 'linear-gradient(135deg, #f6d365 0%, #fda085 100%)' },
];

const DEFAULT_BG_SETTING = { type: 'color', id: 'solid-1', color: '#f5f7fa' };

// 显示聊天背景设置面板
function showChatBgSettings() {
    const panel = document.getElementById('bg-setting-panel');
    if (panel) {
        initBgSettingUI();
        panel.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
    }
}

// 关闭背景设置面板
function closeBgSetting() {
    const panel = document.getElementById('bg-setting-panel');
    if (panel) {
        panel.classList.add('hidden');
        document.body.style.overflow = '';
    }
}

// 恢复默认背景
function resetBg() {
    saveChatBgSetting(DEFAULT_BG_SETTING);
    initBgSettingUI();
    updateChatBgPreview(DEFAULT_BG_SETTING);
}

// 选择背景
function selectBg(bgId, type, value) {
    let bgSetting;
    if (type === 'custom') {
        const customBg = getCustomBg();
        if (!customBg) return;
        bgSetting = { type: 'custom', id: 'custom', url: customBg };
    } else if (type === 'color') {
        bgSetting = { type: 'color', id: bgId, color: value };
    } else {
        bgSetting = { type: 'gradient', id: bgId, gradient: value };
    }

    saveChatBgSetting(bgSetting);
    initBgSettingUI();
    updateChatBgPreview(bgSetting);
}

// 上传背景（内嵌版本）
function uploadBgInline(input) {
    const file = input.files[0];
    if (!file) return;

    if (!file.type.startsWith('image/')) {
        alert('请选择图片文件');
        return;
    }

    const reader = new FileReader();
    reader.onload = function(e) {
        const dataUrl = e.target.result;
        localStorage.setItem('chat_bg_custom', dataUrl);
        const bgSetting = { type: 'custom', id: 'custom', url: dataUrl };
        saveChatBgSetting(bgSetting);
        initBgSettingUI();
        updateChatBgPreview(bgSetting);
    };
    reader.readAsDataURL(file);
}

// 应用背景并关闭
function applyBgAndClose() {
    const bgSetting = getChatBgSetting();
    if (bgSetting) {
        applyChatBg(bgSetting);
    }
    closeBgSetting();
}

// 更新预览区域
function updateChatBgPreview(bgSetting) {
    const chatBg = document.getElementById('bg-chat-preview');
    if (!chatBg) return;

    if (bgSetting.type === 'color') {
        chatBg.style.background = bgSetting.color;
    } else if (bgSetting.type === 'gradient') {
        chatBg.style.background = bgSetting.gradient;
    } else if (bgSetting.type === 'custom' && bgSetting.url) {
        chatBg.style.background = `url(${bgSetting.url}) center/cover`;
    } else {
        chatBg.style.background = DEFAULT_BG_SETTING.color;
    }
}

// 初始化背景设置UI
function initBgSettingUI() {
    const currentBg = getChatBgSetting();

    // 渲染纯色
    const solidList = document.getElementById('bg-solid-list');
    if (solidList) {
        solidList.innerHTML = BG_SOLID_COLORS.map(bg => `
            <div class="bg-color-item ${currentBg && currentBg.id === bg.id ? 'selected' : ''}"
                 style="background-color: ${bg.color}"
                 onclick="selectBg('${bg.id}', 'color', '${bg.color}')"
                 title="${bg.name}">
            </div>
        `).join('');
    }

    // 渲染渐变
    const gradientList = document.getElementById('bg-gradient-list');
    if (gradientList) {
        gradientList.innerHTML = BG_GRADIENT_COLORS.map(bg => `
            <div class="bg-color-item ${currentBg && currentBg.id === bg.id ? 'selected' : ''}"
                 style="background: ${bg.gradient}"
                 onclick="selectBg('${bg.id}', 'gradient', '${bg.gradient.replace(/'/g, "\\'")}')"
                 title="${bg.name}">
            </div>
        `).join('');
    }

    // 更新名称
    const bgNameDisplay = document.getElementById('bg-name-display');
    if (bgNameDisplay) {
        if (currentBg && currentBg.type === 'color') {
            const item = BG_SOLID_COLORS.find(c => c.id === currentBg.id);
            bgNameDisplay.textContent = item ? item.name : '纯色';
        } else if (currentBg && currentBg.type === 'gradient') {
            const item = BG_GRADIENT_COLORS.find(c => c.id === currentBg.id);
            bgNameDisplay.textContent = item ? item.name : '渐变';
        } else {
            bgNameDisplay.textContent = '自定义图片';
        }
    }

    // 更新自定义预览
    const customPreview = document.getElementById('bg-custom-preview');
    const customImg = document.getElementById('bg-custom-img');
    const customBg = getCustomBg();

    if (customBg) {
        customPreview.classList.add('show');
        customImg.src = customBg;
        if (currentBg && currentBg.type === 'custom') {
            customPreview.classList.add('active');
        } else {
            customPreview.classList.remove('active');
        }
    } else {
        customPreview.classList.remove('show');
        customPreview.classList.remove('active');
    }

    // 更新预览
    updateChatBgPreview(currentBg);
}

// 获取聊天背景设置
function getChatBgSetting() {
    const saved = localStorage.getItem(STORAGE_KEYS.CHAT_BG);
    if (saved) {
        try {
            return JSON.parse(saved);
        } catch (e) {
            return DEFAULT_BG_SETTING;
        }
    }
    return DEFAULT_BG_SETTING;
}

// 保存聊天背景设置
function saveChatBgSetting(bgSetting) {
    localStorage.setItem(STORAGE_KEYS.CHAT_BG, JSON.stringify(bgSetting));
}

// 获取自定义背景
function getCustomBg() {
    return localStorage.getItem('chat_bg_custom');
}

// 应用聊天背景
function applyChatBg(bgSetting) {
    const chatContent = document.getElementById('chat-content');
    if (!chatContent) return;

    chatContent.style.background = '';
    chatContent.style.backgroundImage = '';

    if (bgSetting.type === 'color') {
        chatContent.style.backgroundColor = bgSetting.color;
    } else if (bgSetting.type === 'gradient') {
        chatContent.style.background = bgSetting.gradient;
    } else if (bgSetting.type === 'custom' && bgSetting.url) {
        chatContent.style.backgroundImage = `url(${bgSetting.url})`;
        chatContent.style.backgroundSize = 'cover';
        chatContent.style.backgroundPosition = 'center';
    }
}

// 页面加载时应用聊天背景
document.addEventListener('DOMContentLoaded', function() {
    const bgSetting = getChatBgSetting();
    if (bgSetting) {
        applyChatBg(bgSetting);
    }
});
