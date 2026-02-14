package server

func adminPanelHTML(role string, open bool, activeTab string) string {
	ownerOnly := ""
	if role == "owner" {
		ownerOnly = `
        <option value="admin">Admin</option>
        <option value="owner">Owner</option>`
	}

	display := "none"
	if open {
		display = "block"
	}

	autoLoad := ""
	if open {
		autoLoad = `
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', function() { loadTab('` + activeTab + `'); });
} else {
  loadTab('` + activeTab + `');
}`
	}

	tabActive := func(name string) string {
		if name == activeTab {
			return " active"
		}
		return ""
	}

	return `
<!-- Admin Panel -->
<div id="admin-panel" class="admin-card" style="display:` + display + `">
  <div class="admin-header">
    <h2>Admin</h2>
    <a href="/" class="admin-close">&times;</a>
  </div>
  <div class="admin-tabs">
    <a href="/?admin&tab=users" class="admin-tab` + tabActive("users") + `" data-tab="users">Users</a>
    <a href="/?admin&tab=services" class="admin-tab` + tabActive("services") + `" data-tab="services">Services</a>
    <a href="/?admin&tab=access" class="admin-tab` + tabActive("access") + `" data-tab="access">Access</a>
  </div>
  <div id="admin-content" class="admin-body">
  </div>
</div>

<style>
.admin-card {
  background: #1e293b;
  border-radius: 12px;
  max-width: 800px;
  margin: 0 auto 1.5rem;
  box-shadow: 0 4px 24px rgba(0,0,0,0.3);
  overflow: hidden;
}
.admin-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #334155;
}
.admin-header h2 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #f8fafc;
  margin: 0;
}
.admin-close {
  color: #64748b;
  font-size: 0.875rem;
  text-decoration: none;
  width: 1.75rem;
  height: 1.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1.5px solid #475569;
  border-radius: 50%;
  transition: color 0.15s, border-color 0.15s, background 0.15s;
}
.admin-close:hover { color: #fff; border-color: #f97316; background: #f97316; }
.admin-tabs {
  display: flex;
  border-bottom: 1px solid #334155;
}
.admin-tab {
  background: none;
  border: none;
  color: #94a3b8;
  padding: 0.75rem 1.25rem;
  font-size: 0.875rem;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: color 0.15s, border-color 0.15s;
  text-decoration: none;
}
.admin-tab:hover { color: #e2e8f0; }
.admin-tab.active { color: #3b82f6; border-bottom-color: #3b82f6; }
.admin-body {
  padding: 1.25rem 1.5rem;
  overflow-x: auto;
}
.admin-tbl { width:100%;border-collapse:collapse;font-size:0.8125rem; }
.admin-tbl th { text-align:left;color:#94a3b8;font-weight:500;padding:0.5rem 0.75rem;border-bottom:1px solid #334155; }
.admin-tbl td { padding:0.5rem 0.75rem;color:#e2e8f0;border-bottom:1px solid #1e293b; }
.admin-tbl tr:hover td { background:#263044; }
.admin-input {
  background:#0f172a;border:1px solid #334155;border-radius:6px;color:#f8fafc;padding:0.375rem 0.625rem;
  font-size:0.8125rem;outline:none;transition:border-color 0.15s;
}
.admin-input:focus { border-color:#3b82f6; }
.admin-select {
  background:#0f172a;border:1px solid #334155;border-radius:6px;color:#f8fafc;padding:0.375rem 0.625rem;
  font-size:0.8125rem;outline:none;
}
.admin-btn {
  background:#3b82f6;color:#fff;border:none;border-radius:6px;padding:0.375rem 0.75rem;
  font-size:0.8125rem;cursor:pointer;transition:background 0.15s;
}
.admin-btn:hover { background:#2563eb; }
.admin-btn-danger {
  background:#dc2626;color:#fff;border:none;border-radius:6px;padding:0.25rem 0.5rem;
  font-size:0.75rem;cursor:pointer;transition:background 0.15s;
}
.admin-btn-danger:hover { background:#b91c1c; }
.admin-form { display:flex;gap:0.5rem;align-items:center;margin-top:1rem;flex-wrap:wrap; }
.admin-msg { font-size:0.8125rem;padding:0.5rem;border-radius:6px;margin-bottom:0.75rem; }
.admin-msg-ok { background:#14532d;color:#86efac; }
.admin-msg-err { background:#7f1d1d;color:#fca5a5; }
.access-check { width:18px;height:18px;cursor:pointer;accent-color:#3b82f6; }
</style>

<script>
const ROLE = '` + role + `';
let currentTab = 'users';
let adminData = { users: [], services: [], grants: [] };

async function api(method, path, body) {
  var opts = { method: method, headers: { 'Content-Type': 'application/json' } };
  if (body) opts.body = JSON.stringify(body);
  var r = await fetch('/admin/api' + path, opts);
  if (r.status === 204) return null;
  var data = await r.json();
  if (!r.ok) throw new Error(data.error || 'request failed');
  return data;
}

async function loadTab(tab) {
  var el = document.getElementById('admin-content');
  try {
    if (tab === 'users') {
      adminData.users = await api('GET', '/users');
      renderUsers(el);
    } else if (tab === 'services') {
      adminData.services = await api('GET', '/services');
      renderServices(el);
    } else if (tab === 'access') {
      adminData.users = await api('GET', '/users');
      adminData.services = await api('GET', '/services');
      adminData.grants = await api('GET', '/grants');
      renderAccess(el);
    }
  } catch (e) {
    el.innerHTML = '<div class="admin-msg admin-msg-err">' + esc(e.message) + '</div>';
  }
}

function esc(s) {
  var d = document.createElement('div');
  d.textContent = s || '';
  return d.innerHTML;
}

function renderUsers(el) {
  var html = '<table class="admin-tbl"><thead><tr><th>Handle</th><th>Username</th><th>Role</th><th>DID</th><th></th></tr></thead><tbody>';
  for (var i = 0; i < adminData.users.length; i++) {
    var u = adminData.users[i];
    var canChangeRole = ROLE === 'owner';
    var usernameCell = '<input class="admin-input" style="width:90px;font-size:0.75rem" value="' + esc(u.username || '') + '" onchange="updateUsername(' + u.id + ',this.value)">';
    var roleCell = canChangeRole
      ? '<select class="admin-select" onchange="updateRole(' + u.id + ',this.value)">' +
        '<option value="user"' + (u.role==='user'?' selected':'') + '>User</option>' +
        '<option value="admin"' + (u.role==='admin'?' selected':'') + '>Admin</option>' +
        '<option value="owner"' + (u.role==='owner'?' selected':'') + '>Owner</option></select>'
      : esc(u.role);
    var del = '<button class="admin-btn-danger" onclick="deleteUser(' + u.id + ')">Delete</button>';
    html += '<tr><td>' + esc(u.handle || '(no handle)') + '</td><td>' + usernameCell + '</td><td>' + roleCell + '</td><td style="font-size:0.75rem;color:#64748b;max-width:200px;overflow:hidden;text-overflow:ellipsis">' + esc(u.did) + '</td><td>' + del + '</td></tr>';
  }
  html += '</tbody></table>';
  html += '<div class="admin-form">' +
    '<input class="admin-input" id="add-handle" placeholder="handle" style="flex:1;min-width:150px">' +
    '<input class="admin-input" id="add-username" placeholder="username" style="width:90px">' +
    '<select class="admin-select" id="add-role"><option value="user">User</option>` + ownerOnly + `</select>' +
    '<button class="admin-btn" onclick="addUser()">Add</button></div>';
  html += '<div id="users-msg"></div>';
  el.innerHTML = html;
}

async function addUser() {
  var handle = document.getElementById('add-handle').value.trim();
  var username = document.getElementById('add-username').value.trim();
  var role = document.getElementById('add-role').value;
  var msg = document.getElementById('users-msg');
  if (!handle) return;
  try {
    await api('POST', '/users', { handle: handle, role: role, username: username });
    document.getElementById('add-handle').value = '';
    document.getElementById('add-username').value = '';
    msg.className = 'admin-msg admin-msg-ok'; msg.textContent = 'User added';
    loadTab('users');
  } catch (e) {
    msg.className = 'admin-msg admin-msg-err'; msg.textContent = e.message;
  }
}

async function updateUsername(id, username) {
  var msg = document.getElementById('users-msg');
  try {
    await api('PUT', '/users/' + id + '/username', { username: username });
    msg.className = 'admin-msg admin-msg-ok'; msg.textContent = 'Username updated';
    setTimeout(function() { msg.className = ''; msg.textContent = ''; }, 1500);
  } catch (e) {
    msg.className = 'admin-msg admin-msg-err'; msg.textContent = e.message;
    loadTab('users');
  }
}

async function updateRole(id, role) {
  try {
    await api('PUT', '/users/' + id + '/role', { role: role });
  } catch (e) {
    alert(e.message);
    loadTab('users');
  }
}

async function deleteUser(id) {
  if (!confirm('Delete this user?')) return;
  try {
    await api('DELETE', '/users/' + id);
    loadTab('users');
  } catch (e) { alert(e.message); }
}

function renderServices(el) {
  var html = '<table class="admin-tbl"><thead><tr><th>Name</th><th>Slug</th><th>URL</th><th>Admin Role</th><th></th></tr></thead><tbody>';
  for (var i = 0; i < adminData.services.length; i++) {
    var s = adminData.services[i];
    html += '<tr><td>' + esc(s.name) + '</td><td style="color:#64748b">' + esc(s.slug) + '</td><td style="font-size:0.75rem;color:#64748b">' + esc(s.url) + '</td>' +
      '<td><input class="admin-input" style="width:70px;font-size:0.75rem" value="' + esc(s.admin_role) + '" onchange="updateServiceAdminRole(' + s.id + ',this.value)"></td>' +
      '<td><button class="admin-btn-danger" onclick="deleteService(' + s.id + ')">Delete</button></td></tr>';
  }
  html += '</tbody></table>';
  html += '<div class="admin-form">' +
    '<input class="admin-input" id="svc-name" placeholder="Name" style="width:100px">' +
    '<input class="admin-input" id="svc-slug" placeholder="slug" style="width:80px">' +
    '<input class="admin-input" id="svc-url" placeholder="https://..." style="flex:1;min-width:130px">' +
    '<input class="admin-input" id="svc-desc" placeholder="Description" style="width:110px">' +
    '<input class="admin-input" id="svc-admin-role" placeholder="admin" style="width:70px">' +
    '<button class="admin-btn" onclick="addService()">Add</button></div>';
  html += '<div id="services-msg"></div>';
  el.innerHTML = html;
}

async function addService() {
  var name = document.getElementById('svc-name').value.trim();
  var slug = document.getElementById('svc-slug').value.trim();
  var url = document.getElementById('svc-url').value.trim();
  var desc = document.getElementById('svc-desc').value.trim();
  var adminRole = document.getElementById('svc-admin-role').value.trim() || 'admin';
  var msg = document.getElementById('services-msg');
  if (!name || !slug || !url) { msg.className = 'admin-msg admin-msg-err'; msg.textContent = 'Name, slug, and URL required'; return; }
  try {
    await api('POST', '/services', { name: name, slug: slug, url: url, description: desc, icon_url: '', admin_role: adminRole });
    document.getElementById('svc-name').value = '';
    document.getElementById('svc-slug').value = '';
    document.getElementById('svc-url').value = '';
    document.getElementById('svc-desc').value = '';
    document.getElementById('svc-admin-role').value = '';
    msg.className = 'admin-msg admin-msg-ok'; msg.textContent = 'Service added';
    loadTab('services');
  } catch (e) {
    msg.className = 'admin-msg admin-msg-err'; msg.textContent = e.message;
  }
}

async function updateServiceAdminRole(id, adminRole) {
  var svc = null;
  for (var i = 0; i < adminData.services.length; i++) {
    if (adminData.services[i].id === id) { svc = adminData.services[i]; break; }
  }
  if (!svc) return;
  var msg = document.getElementById('services-msg');
  try {
    await api('PUT', '/services/' + id, { name: svc.name, description: svc.description, url: svc.url, icon_url: svc.icon_url, admin_role: adminRole });
    svc.admin_role = adminRole;
    msg.className = 'admin-msg admin-msg-ok'; msg.textContent = 'Admin role updated';
    setTimeout(function() { msg.className = ''; msg.textContent = ''; }, 1500);
  } catch (e) {
    msg.className = 'admin-msg admin-msg-err'; msg.textContent = e.message;
  }
}

async function deleteService(id) {
  if (!confirm('Delete this service? Grants will also be removed.')) return;
  try {
    await api('DELETE', '/services/' + id);
    loadTab('services');
  } catch (e) { alert(e.message); }
}

function renderAccess(el) {
  var users = adminData.users;
  var services = adminData.services;
  var grantMap = {};
  for (var i = 0; i < adminData.grants.length; i++) {
    var g = adminData.grants[i];
    grantMap[g.user_id + ':' + g.service_id] = g;
  }

  var html = '<table class="admin-tbl"><thead><tr><th>User</th>';
  for (var i = 0; i < services.length; i++) {
    html += '<th style="text-align:center;font-size:0.75rem">' + esc(services[i].name) + '</th>';
  }
  html += '</tr></thead><tbody>';
  for (var i = 0; i < users.length; i++) {
    var u = users[i];
    html += '<tr><td>' + esc(u.handle || u.did) + '</td>';
    for (var j = 0; j < services.length; j++) {
      var s = services[j];
      var key = u.id + ':' + s.id;
      var grant = grantMap[key];
      var checked = grant ? ' checked' : '';
      var role = grant ? grant.role : 'user';
      html += '<td style="text-align:center">' +
        '<input type="checkbox" class="access-check"' + checked +
        ' onchange="toggleGrant(' + u.id + ',' + s.id + ',this.checked)">' +
        '<br><input class="admin-input" style="width:60px;font-size:0.6875rem;margin-top:2px;text-align:center" ' +
        'value="' + esc(role) + '" ' +
        'onchange="updateGrantRole(' + u.id + ',' + s.id + ',this.value)"' +
        (grant ? '' : ' disabled') + '></td>';
    }
    html += '</tr>';
  }
  html += '</tbody></table>';
  html += '<div id="access-msg"></div>';
  el.innerHTML = html;
}

async function toggleGrant(userId, serviceId, checked) {
  var msg = document.getElementById('access-msg');
  try {
    if (checked) {
      await api('POST', '/grants', { user_id: userId, service_id: serviceId, role: 'user' });
    } else {
      var grant = null;
      for (var i = 0; i < adminData.grants.length; i++) {
        var g = adminData.grants[i];
        if (g.user_id === userId && g.service_id === serviceId) { grant = g; break; }
      }
      if (grant) {
        await api('DELETE', '/grants/' + grant.id);
      }
    }
    adminData.grants = await api('GET', '/grants');
    renderAccess(document.getElementById('admin-content'));
    msg.className = ''; msg.textContent = '';
  } catch (e) {
    msg.className = 'admin-msg admin-msg-err'; msg.textContent = e.message;
    loadTab('access');
  }
}

async function updateGrantRole(userId, serviceId, role) {
  var msg = document.getElementById('access-msg');
  try {
    await api('POST', '/grants', { user_id: userId, service_id: serviceId, role: role });
    adminData.grants = await api('GET', '/grants');
    msg.className = 'admin-msg admin-msg-ok'; msg.textContent = 'Role updated';
    setTimeout(function() { msg.className = ''; msg.textContent = ''; }, 1500);
  } catch (e) {
    msg.className = 'admin-msg admin-msg-err'; msg.textContent = e.message;
  }
}

` + autoLoad + `
</script>`
}
