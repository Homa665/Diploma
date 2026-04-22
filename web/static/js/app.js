'use strict';

// ===== SVG Icons =====
const ICONS = {
  home: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12l8.954-8.955a1.126 1.126 0 011.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25"/></svg>',
  search: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z"/></svg>',
  plus: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/></svg>',
  chat: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z"/></svg>',
  bell: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0"/></svg>',
  user: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z"/></svg>',
  heart: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z"/></svg>',
  comment: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 20.25c4.97 0 9-3.694 9-8.25s-4.03-8.25-9-8.25S3 7.444 3 12c0 2.104.859 4.023 2.273 5.48.432.447.74 1.04.586 1.641a4.483 4.483 0 01-.923 1.785A5.969 5.969 0 006 21c1.282 0 2.47-.402 3.445-1.087.81.22 1.668.337 2.555.337z"/></svg>',
  share: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5"/></svg>',
  star: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M11.48 3.499a.562.562 0 011.04 0l2.125 5.111a.563.563 0 00.475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 00-.182.557l1.285 5.385a.562.562 0 01-.84.61l-4.725-2.885a.563.563 0 00-.586 0L6.982 20.54a.562.562 0 01-.84-.61l1.285-5.386a.562.562 0 00-.182-.557l-4.204-3.602a.563.563 0 01.321-.988l5.518-.442a.563.563 0 00.475-.345L11.48 3.5z"/></svg>',
  eye: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z"/><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>',
  settings: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z"/><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>',
  close: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg>',
  trash: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"/></svg>',
  edit: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487zm0 0L19.5 7.125"/></svg>',
  flag: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 3v1.5M3 21v-6m0 0l2.77-.693a9 9 0 016.208.682l.108.054a9 9 0 006.086.71l3.114-.732a48.524 48.524 0 01-.005-10.499l-3.11.732a9 9 0 01-6.085-.711l-.108-.054a9 9 0 00-6.208-.682L3 4.5M3 15V4.5"/></svg>',
  team: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z"/></svg>',
  pin: '<svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 24 24"><path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/></svg>',
  document: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z"/></svg>',
  chart: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z"/></svg>',
  shield: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"/></svg>',
};

// ===== API Helper =====
async function api(url, options = {}) {
  const defaults = {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
  };
  if (options.body instanceof FormData) {
    delete defaults.headers['Content-Type'];
  }
  const resp = await fetch(url, { ...defaults, ...options });
  const data = await resp.json();
  if (!resp.ok && data.error) {
    throw new Error(data.error);
  }
  return data;
}

function formData(obj) {
  return new URLSearchParams(obj).toString();
}

// ===== Toast =====
function showToast(msg, type = 'success') {
  const existing = document.querySelector('.toast');
  if (existing) existing.remove();
  const el = document.createElement('div');
  el.className = 'toast';
  el.style.cssText = `
    position:fixed;bottom:20px;left:50%;transform:translateX(-50%);
    padding:12px 24px;border-radius:8px;font-size:.85rem;font-weight:500;
    z-index:3000;animation:fadeInUp .3s;color:#fff;font-family:var(--font);
    background:${type === 'error' ? 'var(--danger)' : type === 'info' ? '#1e40af' : '#16a34a'};
  `;
  el.textContent = msg;
  document.body.appendChild(el);
  setTimeout(() => el.remove(), 3000);
}

// ===== Notification polling =====
let notifInterval = null;
let lastUnreadCount = -1;
let lastNotifIds = new Set();

function showNotifToast(notification) {
  const el = document.createElement('div');
  el.className = 'notif-toast';
  el.style.cssText = `
    position:fixed;top:20px;right:20px;max-width:360px;padding:14px 18px;
    border-radius:12px;background:var(--card);border:1px solid var(--border-light);
    box-shadow:0 8px 32px rgba(0,0,0,0.18);z-index:4000;
    animation:slideInRight .35s ease;cursor:pointer;font-family:var(--font);
  `;
  const icon = notification.type === 'message' ? '💬' : notification.type === 'follow' ? '👤' : notification.type === 'rating' ? '⭐' : notification.type === 'comment' ? '💬' : '🔔';
  el.innerHTML = `
    <div style="display:flex;align-items:flex-start;gap:10px">
      <span style="font-size:1.3rem">${icon}</span>
      <div style="flex:1;min-width:0">
        <div style="font-weight:600;font-size:.82rem;color:var(--text);margin-bottom:2px">Новое уведомление</div>
        <div style="font-size:.8rem;color:var(--text-muted);line-height:1.4;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${escapeHtml(notification.content)}</div>
      </div>
      <button onclick="this.parentElement.parentElement.remove()" style="background:none;border:none;color:var(--text-muted);cursor:pointer;font-size:1.1rem;padding:0;line-height:1">&times;</button>
    </div>
  `;
  if (notification.link) {
    el.addEventListener('click', (e) => {
      if (e.target.tagName !== 'BUTTON') window.location.href = notification.link;
    });
  }
  document.body.appendChild(el);
  setTimeout(() => { if (el.parentNode) el.style.animation = 'slideOutRight .3s ease forwards'; setTimeout(() => el.remove(), 300); }, 5000);
}

function startNotifPolling() {
  async function poll() {
    try {
      const data = await api('/api/notifications');
      const badge = document.getElementById('notif-badge');
      if (badge) {
        if (data.unread_count > 0) {
          badge.textContent = data.unread_count;
          badge.classList.remove('hidden');
        } else {
          badge.classList.add('hidden');
        }
      }
      // Show toast for new notifications
      if (lastUnreadCount >= 0 && data.notifications) {
        data.notifications.forEach(n => {
          if (!n.is_read && !lastNotifIds.has(n.id)) {
            showNotifToast(n);
          }
        });
      }
      // Track known notification IDs
      lastNotifIds = new Set();
      if (data.notifications) {
        data.notifications.forEach(n => lastNotifIds.add(n.id));
      }
      lastUnreadCount = data.unread_count || 0;
    } catch (e) {}
  }
  poll();
  notifInterval = setInterval(poll, 15000);
}

// ===== Chat polling =====
let chatInterval = null;
function startChatPolling(partnerID, userID) {
  const container = document.getElementById('chat-messages');
  if (!container) return;

  async function poll() {
    try {
      const messages = await api('/api/messages?partner_id=' + partnerID);
      container.innerHTML = '';
      messages.forEach(m => {
        const div = document.createElement('div');
        div.className = 'chat-msg ' + (m.sender_id == userID ? 'sent' : 'received');
        const time = new Date(m.created_at);
        const timeStr = time.getHours().toString().padStart(2, '0') + ':' + time.getMinutes().toString().padStart(2, '0');
        div.innerHTML = `
          <div>${renderChatContent(m.content)}</div>
          <div class="chat-msg-time">${timeStr}</div>
        `;
        container.appendChild(div);
      });
      container.scrollTop = container.scrollHeight;
    } catch (e) {}
  }

  chatInterval = setInterval(poll, 3000);
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function renderChatContent(text) {
  const safe = escapeHtml(text);
  return safe.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" style="color:var(--primary);text-decoration:underline">$1</a>');
}

// ===== Modal =====
function openModal(id) {
  const el = document.getElementById(id);
  if (el) el.classList.add('show');
}
function closeModal(id) {
  const el = document.getElementById(id);
  if (el) el.classList.remove('show');
}

// ===== Tab switching =====
function initTabs() {
  document.querySelectorAll('[data-tab-group]').forEach(group => {
    const groupName = group.dataset.tabGroup;
    group.querySelectorAll('[data-tab]').forEach(tab => {
      tab.addEventListener('click', () => {
        group.querySelectorAll('[data-tab]').forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        document.querySelectorAll(`[data-tab-content="${groupName}"]`).forEach(c => c.classList.remove('active'));
        const target = document.getElementById(tab.dataset.tab);
        if (target) target.classList.add('active');
      });
    });
  });
}

// ===== Infinite scroll =====
function initInfiniteScroll() {
  const container = document.getElementById('feed-posts');
  if (!container) return;

  let page = 2;
  let loading = false;
  let hasMore = true;
  const category = new URLSearchParams(window.location.search).get('category') || '';

  window.addEventListener('scroll', async () => {
    if (loading || !hasMore) return;
    if (window.innerHeight + window.scrollY >= document.body.offsetHeight - 500) {
      loading = true;
      const spinner = document.createElement('div');
      spinner.className = 'spinner';
      spinner.id = 'load-spinner';
      container.appendChild(spinner);

      try {
        let url = '/api/feed?page=' + page;
        if (category) url += '&category=' + encodeURIComponent(category);
        const data = await api(url);
        document.getElementById('load-spinner')?.remove();

        if (data.posts && data.posts.length > 0) {
          data.posts.forEach(post => {
            container.insertAdjacentHTML('beforeend', renderPostCard(post));
          });
          page++;
          hasMore = data.has_more;
        } else {
          hasMore = false;
        }
      } catch (e) {
        document.getElementById('load-spinner')?.remove();
      }
      loading = false;
    }
  });
}

function renderPostCard(p) {
  const initials = (p.author_nickname || '?')[0].toUpperCase();
  const avatarHtml = p.author_avatar_url
    ? `<img src="${escapeHtml(p.author_avatar_url)}" alt="">`
    : initials;

  let roleHtml = '';
  if (p.author_role === 'premium') roleHtml = '<span class="role-badge premium">PRO</span>';
  else if (p.author_role === 'expert') roleHtml = '<span class="role-badge expert">Expert</span>';
  else if (p.author_role === 'admin') roleHtml = '<span class="role-badge admin">Admin</span>';

  let filesHtml = '';
  if (p.files && p.files.length > 0) {
    const images = p.files.filter(f => f.file_type && f.file_type.startsWith('image/'));
    if (images.length > 0) {
      const cls = images.length === 1 ? 'single' : 'multi';
      filesHtml = `<div class="post-images ${cls}">` +
        images.slice(0, 4).map(f => `<img src="${escapeHtml(f.file_path)}" alt="">`).join('') +
        `</div>`;
    }
  }

  const tags = p.tags ? p.tags.split(',').filter(t => t.trim()).map(t =>
    `<span class="post-tag">${escapeHtml(t.trim())}</span>`
  ).join('') : '';

  const desc = p.description.length > 200 ? p.description.substring(0, 200) + '...' : p.description;
  const rating = p.avg_rating ? parseFloat(p.avg_rating).toFixed(1) : '0';

  return `<div class="card">
    <div class="post-header">
      <a href="/profile/${p.author_id}" class="post-avatar">${avatarHtml}</a>
      <div class="post-meta">
        <div class="post-author">
          <a href="/profile/${p.author_id}" style="color:inherit">${escapeHtml(p.author_nickname)}</a>
          ${roleHtml}
          ${p.category ? `<span class="category-badge">${escapeHtml(p.category)}</span>` : ''}
        </div>
      </div>
    </div>
    <div class="post-title"><a href="/post/${p.id}">${escapeHtml(p.title)}</a></div>
    <div class="post-desc">${escapeHtml(desc)}</div>
    ${tags ? `<div class="post-tags">${tags}</div>` : ''}
    ${filesHtml}
    <div class="post-stats">
      <span>${ICONS.star} ${rating}</span>
      <span>${ICONS.comment} ${p.comment_count || 0}</span>
      <span>${ICONS.eye} ${p.view_count || 0}</span>
    </div>
  </div>`;
}

// ===== Form handlers =====
function initForms() {
  // Auth forms
  document.querySelectorAll('form[data-api]').forEach(form => {
    form.addEventListener('submit', async e => {
      e.preventDefault();
      const url = form.dataset.api;
      const fd = new FormData(form);
      const btn = form.querySelector('button[type="submit"]');
      if (btn) btn.disabled = true;

      try {
        let resp;
        if (form.enctype === 'multipart/form-data') {
          resp = await api(url, { method: 'POST', body: fd });
        } else {
          const params = new URLSearchParams(fd).toString();
          resp = await api(url, { method: 'POST', body: params });
        }
        if (resp.redirect) {
          window.location.href = resp.redirect;
        } else if (resp.success) {
          showToast('Успешно!');
          if (form.dataset.reload === 'true') {
            setTimeout(() => window.location.reload(), 500);
          }
        }
      } catch (err) {
        showToast(err.message, 'error');
      } finally {
        if (btn) btn.disabled = false;
      }
    });
  });

  // Action buttons
  document.querySelectorAll('[data-action]').forEach(btn => {
    btn.addEventListener('click', async e => {
      e.preventDefault();
      const action = btn.dataset.action;
      const confirm_msg = btn.dataset.confirm;
      if (confirm_msg && !confirm(confirm_msg)) return;

      btn.disabled = true;
      try {
        const body = btn.dataset.body || '';
        const resp = await api(action, { method: 'POST', body });
        if (resp.redirect) {
          window.location.href = resp.redirect;
        } else if (resp.success) {
          showToast('Готово!');
          if (btn.dataset.reload === 'true') {
            setTimeout(() => window.location.reload(), 500);
          }
        }
      } catch (err) {
        showToast(err.message, 'error');
      } finally {
        btn.disabled = false;
      }
    });
  });
}

// ===== Chat send =====
function initChatSend() {
  const form = document.getElementById('chat-send-form');
  if (!form) return;

  form.addEventListener('submit', async e => {
    e.preventDefault();
    const input = form.querySelector('input[name="content"]');
    const receiverID = form.querySelector('input[name="receiver_id"]').value;
    const content = input.value.trim();
    if (!content) return;

    try {
      const resp = await api('/api/messages/send', {
        method: 'POST',
        body: formData({ receiver_id: receiverID, content }),
      });
      if (resp.success) {
        const container = document.getElementById('chat-messages');
        const div = document.createElement('div');
        div.className = 'chat-msg sent';
        div.innerHTML = `
          <div>${renderChatContent(content)}</div>
          <div class="chat-msg-time">${resp.created_at || ''}</div>
        `;
        container.appendChild(div);
        container.scrollTop = container.scrollHeight;
        input.value = '';
      }
    } catch (err) {
      showToast(err.message, 'error');
    }
  });
}

// ===== Rating stars =====
function initRating() {
  const container = document.getElementById('rating-stars');
  if (!container) return;

  const stars = container.querySelectorAll('.rating-star');
  const scoreInput = document.getElementById('rating-score');

  stars.forEach(star => {
    star.addEventListener('click', () => {
      const val = parseInt(star.dataset.value);
      scoreInput.value = val;
      stars.forEach(s => {
        s.classList.toggle('active', parseInt(s.dataset.value) <= val);
      });
    });
    star.addEventListener('mouseenter', () => {
      const val = parseInt(star.dataset.value);
      stars.forEach(s => {
        s.classList.toggle('active', parseInt(s.dataset.value) <= val);
      });
    });
  });

  container.addEventListener('mouseleave', () => {
    const val = parseInt(scoreInput.value) || 0;
    stars.forEach(s => {
      s.classList.toggle('active', parseInt(s.dataset.value) <= val);
    });
  });
}

// ===== Init =====
document.addEventListener('DOMContentLoaded', () => {
  initForms();
  initTabs();
  initChatSend();
  initRating();
  initInfiniteScroll();

  // Start notification polling if logged in
  if (document.getElementById('notif-badge')) {
    startNotifPolling();
  }

  // Start chat polling if on chat page
  const chatContainer = document.getElementById('chat-messages');
  if (chatContainer) {
    const partnerID = chatContainer.dataset.partnerId;
    const userID = chatContainer.dataset.userId;
    if (partnerID) {
      startChatPolling(partnerID, userID);
    }
  }

  // Scroll to bottom of chat
  if (chatContainer) {
    chatContainer.scrollTop = chatContainer.scrollHeight;
  }
});

// CSS for toast animation
const style = document.createElement('style');
style.textContent = '@keyframes fadeInUp{from{opacity:0;transform:translate(-50%,20px)}to{opacity:1;transform:translate(-50%,0)}}@keyframes slideInRight{from{opacity:0;transform:translateX(100%)}to{opacity:1;transform:translateX(0)}}@keyframes slideOutRight{from{opacity:1;transform:translateX(0)}to{opacity:0;transform:translateX(100%)}}';
document.head.appendChild(style);
