package web

var templates = map[string]string{
"login": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0,maximum-scale=1.0,user-scalable=no">
<title>Dday AC - Login</title>
<meta name="theme-color" content="#1e293b">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="Dday AC">
<link rel="manifest" href="/manifest.json">
<link rel="icon" type="image/svg+xml" href="/static/icon.svg">
<link rel="apple-touch-icon" href="/static/icon.svg">
<link rel="stylesheet" href="/static/style.css">
</head><body class="login-body">
<div class="login-container">
  <div class="login-header">
    <img src="/static/logo.svg" alt="Dday AC" class="login-logo-img">
    <p>Dashboard de Administración</p>
  </div>
  {{if .Error}}<div class="alert alert-error">{{.Error}}</div>{{end}}
  <form method="POST" action="/login">
    <div class="form-group"><label>Usuario</label>
      <input type="text" name="username" required autofocus placeholder="admin"></div>
    <div class="form-group"><label>Contraseña</label>
      <input type="password" name="password" required placeholder="&bull;&bull;&bull;&bull;&bull;&bull;"></div>
    <button type="submit" class="btn btn-primary btn-block">Entrar</button>
  </form>
</div></body></html>`,

"dashboard": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0,maximum-scale=1.0,user-scalable=no">
<title>Dday AC - Dashboard</title>
<meta name="theme-color" content="#1e293b">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="Dday AC">
<link rel="manifest" href="/manifest.json">
<link rel="icon" type="image/svg+xml" href="/static/icon.svg">
<link rel="apple-touch-icon" href="/static/icon.svg">
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar">
  <h2>Dashboard</h2>
  <div style="display:flex;align-items:center;gap:12px;">
    <button type="button" class="push-toggle-btn" onclick="togglePushSubscription()" title="Activar/Desactivar Notificaciones Push">
      <span>🔔</span> Notificaciones
    </button>
    <form method="GET" action="/player" class="quick-search-form" style="display:flex;gap:8px;">
      <input type="text" name="q" placeholder="Buscar jugador por nombre o IP..." class="form-control" style="width:240px;padding:6px 12px;border-radius:6px;background:var(--bg-primary);border:1px solid var(--border);color:var(--text-primary);font-size:13px;">
      <button type="submit" class="btn btn-primary btn-sm">&#128269; Buscar</button>
    </form>
  </div>
</div>
<div class="content">

<div class="stats-grid">
  <div class="stat-card"><div class="stat-icon blue">&#127760;</div><div class="stat-info"><div class="stat-value">{{.ServerCount}}</div><div class="stat-label">Servidores Conectados</div></div></div>
  <div class="stat-card"><div class="stat-icon green">&#128100;</div><div class="stat-info"><div class="stat-value">{{.ClientCount}}</div><div class="stat-label">Jugadores Activos</div></div></div>
  <div class="stat-card"><div class="stat-icon purple">&#128247;</div><div class="stat-info"><div class="stat-value">{{index .Stats "total_screenshots"}}</div><div class="stat-label">Total Screenshots</div></div></div>
  <div class="stat-card"><div class="stat-icon orange">&#9888;</div><div class="stat-info"><div class="stat-value">{{index .Stats "total_violations"}}</div><div class="stat-label">Total Violations</div></div></div>
  <div class="stat-card"><div class="stat-icon purple">&#128737;</div><div class="stat-info"><div class="stat-value">{{index .Stats "total_process_snapshots"}}</div><div class="stat-label">Process Snapshots</div></div></div>
</div>

{{if .DailyActivity}}
<div class="card chart-card" style="margin-bottom: 24px;">
  <div class="card-header"><h3>&#128200; Actividad de los Últimos 7 Días</h3></div>
  <div class="card-body">
    <div style="display:flex; gap:16px; margin-bottom:12px; font-size:13px; flex-wrap:wrap;">
      <span style="display:flex;align-items:center;gap:6px;"><span style="display:inline-block;width:12px;height:12px;background:#3b82f6;border-radius:2px;"></span> Screenshots</span>
      <span style="display:flex;align-items:center;gap:6px;"><span style="display:inline-block;width:12px;height:12px;background:#ef4444;border-radius:2px;"></span> Violaciones</span>
      <span style="display:flex;align-items:center;gap:6px;"><span style="display:inline-block;width:12px;height:12px;background:#a855f7;border-radius:2px;"></span> Process Snapshots</span>
    </div>
    <div class="chart-container" style="display:flex; gap:12px; justify-content:space-between; align-items:flex-end; height:160px; padding:10px 0; border-bottom:1px solid var(--border);">
      {{range .DailyActivity}}
      <div style="flex:1; display:flex; flex-direction:column; align-items:center; height:100%; justify-content:flex-end;">
        <div style="display:flex; gap:4px; align-items:flex-end; height:120px; width:100%; max-width:48px; justify-content:center;">
          <div style="width:10px; background:#3b82f6; border-radius:3px 3px 0 0; min-height:4px; height:{{if gt .Screenshots 0}}{{add 4 (add .Screenshots .Screenshots)}}{{else}}4{{end}}px;" title="{{.Date}}: {{.Screenshots}} screenshots"></div>
          <div style="width:10px; background:#ef4444; border-radius:3px 3px 0 0; min-height:4px; height:{{if gt .Violations 0}}{{add 4 (add .Violations .Violations)}}{{else}}4{{end}}px;" title="{{.Date}}: {{.Violations}} violaciones"></div>
          <div style="width:10px; background:#a855f7; border-radius:3px 3px 0 0; min-height:4px; height:{{if gt .Snapshots 0}}{{add 4 (add .Snapshots .Snapshots)}}{{else}}4{{end}}px;" title="{{.Date}}: {{.Snapshots}} snapshots"></div>
        </div>
        <div style="font-size:11px; color:var(--text-secondary); margin-top:8px;">{{.Date}}</div>
      </div>
      {{end}}
    </div>
  </div>
</div>
{{end}}

<div class="grid-2">
  <div class="card"><div class="card-header"><h3>Resumen y Estado</h3></div><div class="card-body">
    <div class="info-row"><span>Screenshots sin revisar</span><a href="/screenshots?unreviewed=1" class="badge badge-warning" style="text-decoration:none;">{{index .Stats "unreviewed_screenshots"}} pendientes</a></div>
    <div class="info-row"><span>Violations registradas hoy</span><span class="badge badge-danger">{{index .Stats "today_violations"}}</span></div>
    <div class="info-row"><span>Espacio en disco utilizado</span><span class="badge badge-info" id="total-size">{{index .Stats "total_size"}} bytes</span></div>
    <div class="info-row"><span>Mantenimiento y Purga</span><a href="/settings" class="btn btn-sm" style="text-decoration:none;">Ajustes del Sistema &rarr;</a></div>
  </div></div>

  <div class="card"><div class="card-header"><h3>Accesos Rápidos</h3></div><div class="card-body">
    <a href="/screenshots?unreviewed=1" class="quick-link"><span class="ql-icon">&#128247;</span><span>Screenshots pendientes de revisión</span></a>
    <a href="/violations?type=file" class="quick-link"><span class="ql-icon">&#128196;</span><span>Violaciones de archivos modificados</span></a>
    <a href="/violations?type=cvar" class="quick-link"><span class="ql-icon">&#128260;</span><span>Violaciones de cvars bloqueadas</span></a>
    <a href="/blacklist" class="quick-link"><span class="ql-icon">&#128683;</span><span>Administrar Blacklist de Procesos/DLLs/SHA1</span></a>
    <a href="/servers" class="quick-link"><span class="ql-icon">&#127760;</span><span>Monitoreo de servidores activos</span></a>
  </div></div>
</div>

<div class="grid-2" style="margin-top:24px;">
  <div class="card">
    <div class="card-header" style="display:flex;justify-content:space-between;align-items:center;">
      <h3>&#9888; Violaciones Recientes</h3>
      <a href="/violations" class="btn btn-sm">Ver todas &rarr;</a>
    </div>
    <div class="card-body" style="padding:0;">
      {{if .RecentViolations}}
      <div class="table-responsive"><table class="data-table compact">
        <thead><tr><th>Fecha</th><th>Jugador</th><th>Tipo</th><th>Razón</th></tr></thead>
        <tbody>{{range .RecentViolations}}<tr>
          <td>{{.Timestamp.Format "01-02 15:04"}}</td>
          <td><a href="/player?q={{.PlayerName}}" class="player-link">{{.PlayerName}}</a></td>
          <td><span class="badge badge-{{if eq .Type "file"}}warning{{else if eq .Type "cvar"}}danger{{else}}info{{end}}">{{.Type}}</span></td>
          <td style="font-size:12px;" title="{{.Reason}}">{{.Reason}}</td>
        </tr>{{end}}</tbody>
      </table></div>
      {{else}}<div class="empty-state small"><p>No hay violaciones recientes</p></div>{{end}}
    </div>
  </div>

  <div class="card">
    <div class="card-header" style="display:flex;justify-content:space-between;align-items:center;">
      <h3>&#128247; Últimos Screenshots</h3>
      <a href="/screenshots" class="btn btn-sm">Ver todos &rarr;</a>
    </div>
    <div class="card-body">
      {{if .RecentScreenshots}}
      <div style="display:grid; grid-template-columns:repeat(auto-fill, minmax(130px, 1fr)); gap:12px;">
        {{range .RecentScreenshots}}
        <div style="background:var(--bg-primary); border-radius:6px; overflow:hidden; border:1px solid var(--border);">
          <div style="position:relative; cursor:pointer; height:90px; overflow:hidden;" onclick="openLightbox('/screenshots/image/{{.ID}}','{{.PlayerName}}','{{.PlayerIP}}','{{.Timestamp.Format "2006-01-02 15:04:05"}}','{{.ServerAddr}}')">
            <img src="/screenshots/image/{{.ID}}" style="width:100%; height:100%; object-fit:cover;" alt="Thumbnail" loading="lazy">
            {{if not .Reviewed}}<span class="badge badge-new" style="position:absolute; top:4px; right:4px; font-size:10px;">NEW</span>{{end}}
          </div>
          <div style="padding:6px 8px; font-size:11px;">
            <a href="/player?q={{.PlayerName}}" class="player-link" style="display:block; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; font-weight:600;">{{.PlayerName}}</a>
            <div style="color:var(--text-secondary); font-size:10px;">{{.Timestamp.Format "15:04:05"}}</div>
          </div>
        </div>
        {{end}}
      </div>
      {{else}}<div class="empty-state small"><p>No hay capturas recientes</p></div>{{end}}
    </div>
  </div>
</div>

</div></div>

<div id="lightbox" class="lightbox" onclick="closeLightbox()">
  <div class="lightbox-content" onclick="event.stopPropagation()">
    <div class="lightbox-toolbar">
      <button class="lightbox-btn" onclick="zoomLightbox(0.25)" title="Acercar (+)">&#43;</button>
      <button class="lightbox-btn" onclick="zoomLightbox(-0.25)" title="Alejar (-)">&#8722;</button>
      <button class="lightbox-btn" onclick="resetLightboxZoom()" title="Restablecer Zoom (0)">1:1</button>
      <button class="lightbox-btn" onclick="toggleLightboxInvert()" title="Invertir Colores / Chams (I)">&#9680; Invertir</button>
      <a id="lightbox-profile-btn" href="#" class="lightbox-btn" target="_blank" title="Ver Perfil del Jugador">&#128100; Perfil</a>
      <a id="lightbox-download-btn" href="#" class="lightbox-btn" download title="Descargar Imagen">&#128190; Descargar</a>
      <button class="lightbox-close" onclick="closeLightbox()" title="Cerrar (Esc)">&times;</button>
    </div>
    <div class="lightbox-img-wrapper" style="overflow:auto; max-height:80vh; display:flex; justify-content:center; align-items:center;">
      <img id="lightbox-img" src="" alt="Screenshot" style="transition:transform 0.15s ease, filter 0.15s ease; max-width:100%; max-height:78vh; object-fit:contain;">
    </div>
    <div id="lightbox-info" class="lightbox-info"></div>
  </div>
</div>

<script src="/static/app.js"></script>
</body></html>`,

"player": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0,maximum-scale=1.0,user-scalable=no">
<title>Dday AC - Perfil de Jugador</title>
<meta name="theme-color" content="#1e293b">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="Dday AC">
<link rel="manifest" href="/manifest.json">
<link rel="icon" type="image/svg+xml" href="/static/icon.svg">
<link rel="apple-touch-icon" href="/static/icon.svg">
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar">
  <h2>&#128100; Perfil Unificado de Jugador</h2>
  <form method="GET" action="/player" class="quick-search-form" style="display:flex;gap:8px;">
    <input type="text" name="q" value="{{.Query}}" placeholder="Buscar otro jugador o IP..." class="form-control" style="width:260px;padding:6px 12px;border-radius:6px;background:var(--bg-primary);border:1px solid var(--border);color:var(--text-primary);font-size:13px;">
    <button type="submit" class="btn btn-primary btn-sm">&#128269; Buscar</button>
  </form>
</div>
<div class="content">

{{if not .Query}}
<div class="card"><div class="card-body" style="text-align:center; padding:48px 24px;">
  <div style="font-size:48px; margin-bottom:12px;">&#128269;</div>
  <h3>Búsqueda de Jugador</h3>
  <p style="color:var(--text-secondary); max-width:500px; margin:8px auto 20px auto;">
    Ingrese un nombre exacto, alias parcial o dirección IP para consultar todo el historial consolidado de capturas, violaciones y procesos.
  </p>
  <form method="GET" action="/player" style="display:inline-flex; gap:8px;">
    <input type="text" name="q" placeholder="Ej: player, 192.168.1.50..." required style="width:300px; padding:10px 14px; border-radius:6px; background:var(--bg-primary); border:1px solid var(--border); color:var(--text-primary);">
    <button type="submit" class="btn btn-primary">Buscar Perfil</button>
  </form>
</div></div>
{{else if not .Profile}}
<div class="card"><div class="card-body empty-state">
  <div style="font-size:48px; margin-bottom:12px;">&#128533;</div>
  <h3>No se encontraron registros para "{{.Query}}"</h3>
  <p>No hay capturas, violaciones ni snapshots asociados a este nombre o dirección IP.</p>
  <a href="/" class="btn btn-primary" style="margin-top:16px;">Volver al Dashboard</a>
</div></div>
{{else}}

<div class="profile-card">
  <div style="display:flex; align-items:center; gap:16px; flex-wrap:wrap;">
    <div class="profile-avatar">&#128100;</div>
    <div style="flex:1;">
      <h2 style="margin:0 0 6px 0; font-size:24px; color:var(--text-primary);">{{.Profile.PrimaryName}}</h2>
      <div style="display:flex; gap:12px; flex-wrap:wrap; font-size:13px; color:var(--text-secondary);">
        <span>&#128247; {{.Profile.ScreenshotCount}} Capturas ({{.Profile.UnreviewedCount}} pendientes)</span>
        <span>&#9888; <strong style="color:#ef4444;">{{.Profile.ViolationCount}} Violaciones</strong></span>
        <span>&#128737; {{.Profile.ProcessSnapshotsCount}} Snapshots</span>
      </div>
    </div>
  </div>

  <div style="display:grid; grid-template-columns:repeat(auto-fit, minmax(280px, 1fr)); gap:16px; margin-top:20px; padding-top:16px; border-top:1px solid var(--border);">
    <div>
      <div style="font-size:12px; text-transform:uppercase; color:var(--text-secondary); margin-bottom:6px; font-weight:600;">Nombres y Alias Conocidos</div>
      <div class="profile-tags">
        {{range .Profile.Aliases}}
        <span class="profile-tag name">{{.}}</span>
        {{else}}<span style="color:var(--text-secondary); font-size:12px;">Sin alias adicionales</span>{{end}}
      </div>
    </div>
    <div>
      <div style="font-size:12px; text-transform:uppercase; color:var(--text-secondary); margin-bottom:6px; font-weight:600;">Direcciones IP Utilizadas</div>
      <div class="profile-tags">
        {{range .Profile.IPs}}
        <span class="profile-tag ip"><code>{{.}}</code></span>
        {{else}}<span style="color:var(--text-secondary); font-size:12px;">No registradas</span>{{end}}
      </div>
    </div>
    <div>
      <div style="font-size:12px; text-transform:uppercase; color:var(--text-secondary); margin-bottom:6px; font-weight:600;">Servidores Frecuentados</div>
      <div class="profile-tags">
        {{range .Profile.Servers}}
        <span class="profile-tag server">{{.}}</span>
        {{else}}<span style="color:var(--text-secondary); font-size:12px;">No registrados</span>{{end}}
      </div>
    </div>
  </div>
</div>

<div class="card" style="margin-top:24px;">
  <div class="card-header"><h3>&#9888; Historial de Violaciones ({{len .Profile.Violations}})</h3></div>
  <div class="card-body">
    {{if .Profile.Violations}}
    <div class="table-responsive"><table class="data-table">
      <thead><tr><th>Fecha</th><th>Servidor</th><th>IP</th><th>Tipo</th><th>Razón</th></tr></thead>
      <tbody>{{range .Profile.Violations}}<tr>
        <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
        <td>{{.ServerAddr}}</td>
        <td><code>{{.PlayerIP}}</code></td>
        <td><span class="badge badge-{{if eq .Type "file"}}warning{{else if eq .Type "cvar"}}danger{{else}}info{{end}}">{{.Type}}</span></td>
        <td>{{.Reason}}</td>
      </tr>{{end}}</tbody>
    </table></div>
    {{else}}<p style="color:var(--text-secondary); margin:0;">No se registran violaciones para este jugador.</p>{{end}}
  </div>
</div>

<div class="card" style="margin-top:24px;">
  <div class="card-header"><h3>&#128247; Galería de Screenshots ({{len .Profile.Screenshots}})</h3></div>
  <div class="card-body">
    {{if .Profile.Screenshots}}
    <div class="screenshot-grid">{{range .Profile.Screenshots}}
      <div class="screenshot-card {{if .Reviewed}}reviewed{{end}}">
        <div class="screenshot-img" onclick="openLightbox('/screenshots/image/{{.ID}}','{{.PlayerName}}','{{.PlayerIP}}','{{.Timestamp.Format "2006-01-02 15:04:05"}}','{{.ServerAddr}}')">
          <img src="/screenshots/image/{{.ID}}" alt="Screenshot" loading="lazy">
          {{if not .Reviewed}}<span class="badge badge-new">NUEVO</span>{{end}}
        </div>
        <div class="screenshot-info">
          <div class="ss-name">{{.PlayerName}}</div>
          <div class="ss-ip"><code>{{.PlayerIP}}</code></div>
          <div class="ss-date">{{.Timestamp.Format "2006-01-02 15:04"}}</div>
          <div class="ss-server">{{.ServerAddr}}</div>
          {{if not .Reviewed}}
          <form method="POST" action="/screenshots/review" class="review-form">
            <input type="hidden" name="id" value="{{.ID}}">
            <input type="text" name="notes" placeholder="Notas...">
            <button type="submit" class="btn btn-sm btn-success">Marcar revisado</button>
          </form>
          {{else}}<div class="reviewed-badge">Revisado{{if .Notes}}: {{.Notes}}{{end}}</div>{{end}}
        </div>
      </div>
    {{end}}</div>
    {{else}}<p style="color:var(--text-secondary); margin:0;">No hay capturas registradas para este jugador.</p>{{end}}
  </div>
</div>

<div class="card" style="margin-top:24px;">
  <div class="card-header"><h3>&#128737; Snapshots de Procesos ({{len .Profile.ProcessSnapshots}})</h3></div>
  <div class="card-body">
    {{if .Profile.ProcessSnapshots}}
    <div class="table-responsive"><table class="data-table">
      <thead><tr><th>ID</th><th>Servidor</th><th>IP</th><th>Procesos</th><th>Módulos</th><th>Estado</th><th>Fecha</th><th>Acción</th></tr></thead>
      <tbody>{{range .Profile.ProcessSnapshots}}<tr>
        <td>#{{.ID}}</td>
        <td>{{.ServerAddr}}</td>
        <td><code>{{.PlayerIP}}</code></td>
        <td>{{.NumProcesses}}</td>
        <td>{{.NumModules}}</td>
        <td>{{if .Violations}}<span class="badge badge-danger">{{violationsCount .Violations}} violaciones</span>{{else}}<span class="badge badge-success">Clean</span>{{end}}</td>
        <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
        <td><a href="/process-snapshots/{{.ID}}" class="btn btn-sm">Ver Detalle &rarr;</a></td>
      </tr>{{end}}</tbody>
    </table></div>
    {{else}}<p style="color:var(--text-secondary); margin:0;">No hay process snapshots para este jugador.</p>{{end}}
  </div>
</div>

{{end}}

</div></div>

<div id="lightbox" class="lightbox" onclick="closeLightbox()">
  <div class="lightbox-content" onclick="event.stopPropagation()">
    <div class="lightbox-toolbar">
      <button class="lightbox-btn" onclick="zoomLightbox(0.25)" title="Acercar (+)">&#43;</button>
      <button class="lightbox-btn" onclick="zoomLightbox(-0.25)" title="Alejar (-)">&#8722;</button>
      <button class="lightbox-btn" onclick="resetLightboxZoom()" title="Restablecer Zoom (0)">1:1</button>
      <button class="lightbox-btn" onclick="toggleLightboxInvert()" title="Invertir Colores / Chams (I)">&#9680; Invertir</button>
      <a id="lightbox-profile-btn" href="#" class="lightbox-btn" target="_blank" title="Ver Perfil del Jugador">&#128100; Perfil</a>
      <a id="lightbox-download-btn" href="#" class="lightbox-btn" download title="Descargar Imagen">&#128190; Descargar</a>
      <button class="lightbox-close" onclick="closeLightbox()" title="Cerrar (Esc)">&times;</button>
    </div>
    <div class="lightbox-img-wrapper" style="overflow:auto; max-height:80vh; display:flex; justify-content:center; align-items:center;">
      <img id="lightbox-img" src="" alt="Screenshot" style="transition:transform 0.15s ease, filter 0.15s ease; max-width:100%; max-height:78vh; object-fit:contain;">
    </div>
    <div id="lightbox-info" class="lightbox-info"></div>
  </div>
</div>

<script src="/static/app.js"></script>
</body></html>`,

"screenshots": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0,maximum-scale=1.0,user-scalable=no">
<title>Dday AC - Screenshots</title>
<meta name="theme-color" content="#1e293b">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="Dday AC">
<link rel="manifest" href="/manifest.json">
<link rel="icon" type="image/svg+xml" href="/static/icon.svg">
<link rel="apple-touch-icon" href="/static/icon.svg">
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar">
  <h2>Screenshots</h2>
  <a href="/screenshots/export.csv?name={{.PlayerName}}&player={{.PlayerIP}}&server={{.ServerAddr}}&from={{.DateFrom}}&to={{.DateTo}}{{if .Unreviewed}}&unreviewed=1{{end}}" class="btn btn-sm btn-outline" title="Descargar registros filtrados en formato CSV">&#128190; Exportar CSV</a>
</div>
<div class="content">

<div class="card"><div class="card-header"><h3>Filtros de Búsqueda</h3></div><div class="card-body">
  <form method="GET" action="/screenshots" class="filter-form"><div class="form-row">
    <div class="form-group"><label>Nombre Jugador (parcial/exacto)</label><input type="text" name="name" value="{{.PlayerName}}" placeholder="Buscar nombre..."></div>
    <div class="form-group"><label>IP Jugador</label><input type="text" name="player" value="{{.PlayerIP}}" placeholder="Filtrar por IP..."></div>
    <div class="form-group"><label>Servidor</label>
      <select name="server">
        <option value="">Todos los servidores</option>
        {{range .Servers}}
        <option value="{{.}}" {{if eq $.ServerAddr .}}selected{{end}}>{{.}}</option>
        {{end}}
      </select>
    </div>
    <div class="form-group"><label>Desde</label><input type="date" name="from" value="{{.DateFrom}}"></div>
    <div class="form-group"><label>Hasta</label><input type="date" name="to" value="{{.DateTo}}"></div>
    <div class="form-group"><label>Por página</label>
      <select name="per_page">
        <option value="20" {{if eq .PerPage 20}}selected{{end}}>20</option>
        <option value="50" {{if eq .PerPage 50}}selected{{end}}>50</option>
        <option value="100" {{if eq .PerPage 100}}selected{{end}}>100</option>
      </select>
    </div>
    <div class="form-group"><label>&nbsp;</label><label class="checkbox-label"><input type="checkbox" name="unreviewed" value="1" {{if .Unreviewed}}checked{{end}}> Solo sin revisar</label></div>
    <div class="form-group"><label>&nbsp;</label><button type="submit" class="btn btn-primary">Filtrar</button></div>
  </div></form>
</div></div>

<div class="card">
  <div class="card-header" style="display:flex; justify-content:space-between; align-items:center; flex-wrap:wrap; gap:12px;">
    <h3>Screenshots ({{.Total}} total)</h3>
    <div class="bulk-toolbar" style="display:flex; align-items:center; gap:12px;">
      <label style="font-size:13px; display:flex; align-items:center; gap:6px; cursor:pointer;">
        <input type="checkbox" onchange="toggleSelectAllScreenshots(this)"> Seleccionar todo
      </label>
      <button id="bulk-review-btn" type="button" class="btn btn-sm btn-success" onclick="submitBulkReview()" disabled>
        &#10004; Marcar (<span id="selected-count">0</span>) como revisados
      </button>
      <form id="bulk-review-form" method="POST" action="/screenshots/bulk-review" style="display:none;">
        <input type="hidden" name="ids" id="bulk-ids-input">
      </form>
    </div>
  </div>
  <div class="card-body">
{{if .Screenshots}}
<div class="screenshot-grid">{{range .Screenshots}}
  <div class="screenshot-card {{if .Reviewed}}reviewed{{end}}">
    <div style="position:absolute; top:8px; left:8px; z-index:5;">
      <input type="checkbox" class="ss-checkbox" value="{{.ID}}" onchange="updateBulkActionState()" style="transform:scale(1.2); cursor:pointer;">
    </div>
    <div class="screenshot-img" onclick="openLightbox('/screenshots/image/{{.ID}}','{{.PlayerName}}','{{.PlayerIP}}','{{.Timestamp.Format "2006-01-02 15:04:05"}}','{{.ServerAddr}}')">
      <img src="/screenshots/image/{{.ID}}" alt="Screenshot" loading="lazy">
      {{if not .Reviewed}}<span class="badge badge-new">NUEVO</span>{{end}}
    </div>
    <div class="screenshot-info">
      <div class="ss-name"><a href="/player?q={{.PlayerName}}" class="player-link">{{.PlayerName}}</a></div>
      <div class="ss-ip"><code>{{.PlayerIP}}</code></div>
      <div class="ss-date">{{.Timestamp.Format "2006-01-02 15:04"}}</div>
      <div class="ss-server">{{.ServerAddr}}</div>
      {{if not .Reviewed}}
      <form method="POST" action="/screenshots/review" class="review-form">
        <input type="hidden" name="id" value="{{.ID}}">
        <input type="text" name="notes" placeholder="Notas...">
        <button type="submit" class="btn btn-sm btn-success">Revisado</button>
      </form>
      {{else}}<div class="reviewed-badge">Revisado{{if .Notes}}: {{.Notes}}{{end}}</div>{{end}}
    </div>
  </div>
{{end}}</div>

{{if gt .TotalPages 1}}<div class="pagination">
  {{if gt .Page 1}}<a href="?page={{sub .Page 1}}&per_page={{.PerPage}}&name={{.PlayerName}}&player={{.PlayerIP}}&server={{.ServerAddr}}&from={{.DateFrom}}&to={{.DateTo}}{{if .Unreviewed}}&unreviewed=1{{end}}" class="btn btn-sm">Anterior</a>{{end}}
  <span class="page-info">Página {{.Page}} de {{.TotalPages}}</span>
  {{if lt .Page .TotalPages}}<a href="?page={{add .Page 1}}&per_page={{.PerPage}}&name={{.PlayerName}}&player={{.PlayerIP}}&server={{.ServerAddr}}&from={{.DateFrom}}&to={{.DateTo}}{{if .Unreviewed}}&unreviewed=1{{end}}" class="btn btn-sm">Siguiente</a>{{end}}
</div>{{end}}
{{else}}<div class="empty-state"><p>No se encontraron screenshots con los filtros aplicados</p></div>{{end}}
</div></div>

</div></div>

<div id="lightbox" class="lightbox" onclick="closeLightbox()">
  <div class="lightbox-content" onclick="event.stopPropagation()">
    <div class="lightbox-toolbar">
      <button class="lightbox-btn" onclick="zoomLightbox(0.25)" title="Acercar (+)">&#43;</button>
      <button class="lightbox-btn" onclick="zoomLightbox(-0.25)" title="Alejar (-)">&#8722;</button>
      <button class="lightbox-btn" onclick="resetLightboxZoom()" title="Restablecer Zoom (0)">1:1</button>
      <button class="lightbox-btn" onclick="toggleLightboxInvert()" title="Invertir Colores / Chams (I)">&#9680; Invertir</button>
      <a id="lightbox-profile-btn" href="#" class="lightbox-btn" target="_blank" title="Ver Perfil del Jugador">&#128100; Perfil</a>
      <a id="lightbox-download-btn" href="#" class="lightbox-btn" download title="Descargar Imagen">&#128190; Descargar</a>
      <button class="lightbox-close" onclick="closeLightbox()" title="Cerrar (Esc)">&times;</button>
    </div>
    <div class="lightbox-img-wrapper" style="overflow:auto; max-height:80vh; display:flex; justify-content:center; align-items:center;">
      <img id="lightbox-img" src="" alt="Screenshot" style="transition:transform 0.15s ease, filter 0.15s ease; max-width:100%; max-height:78vh; object-fit:contain;">
    </div>
    <div id="lightbox-info" class="lightbox-info"></div>
  </div>
</div>

<script src="/static/app.js"></script>
</body></html>`,

"violations": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0,maximum-scale=1.0,user-scalable=no">
<title>Dday AC - Violations</title>
<meta name="theme-color" content="#1e293b">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="Dday AC">
<link rel="manifest" href="/manifest.json">
<link rel="icon" type="image/svg+xml" href="/static/icon.svg">
<link rel="apple-touch-icon" href="/static/icon.svg">
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar">
  <h2>Violations</h2>
  <div style="display:flex;align-items:center;gap:12px;">
    <button type="button" class="push-toggle-btn" onclick="togglePushSubscription()" title="Activar/Desactivar Notificaciones Push">
      <span>🔔</span> Notificaciones
    </button>
    <a href="/violations/export.csv?name={{.PlayerName}}&player={{.PlayerIP}}&server={{.ServerAddr}}&type={{.Type}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm btn-outline" title="Descargar historial de violaciones en formato CSV">&#128190; Exportar CSV</a>
  </div>
</div>
<div class="content">

<div class="card"><div class="card-header"><h3>Filtros de Búsqueda</h3></div><div class="card-body">
  <form method="GET" action="/violations" class="filter-form"><div class="form-row">
    <div class="form-group"><label>Nombre Jugador (parcial/exacto)</label><input type="text" name="name" value="{{.PlayerName}}" placeholder="Buscar nombre..."></div>
    <div class="form-group"><label>IP Jugador</label><input type="text" name="player" value="{{.PlayerIP}}" placeholder="Filtrar por IP..."></div>
    <div class="form-group"><label>Servidor</label>
      <select name="server">
        <option value="">Todos los servidores</option>
        {{range .Servers}}
        <option value="{{.}}" {{if eq $.ServerAddr .}}selected{{end}}>{{.}}</option>
        {{end}}
      </select>
    </div>
    <div class="form-group"><label>Tipo</label><select name="type">
      <option value="">Todos los tipos</option>
      <option value="file" {{if eq .Type "file"}}selected{{end}}>Archivos</option>
      <option value="cvar" {{if eq .Type "cvar"}}selected{{end}}>Cvars</option>
    </select></div>
    <div class="form-group"><label>Desde</label><input type="date" name="from" value="{{.DateFrom}}"></div>
    <div class="form-group"><label>Hasta</label><input type="date" name="to" value="{{.DateTo}}"></div>
    <div class="form-group"><label>Por página</label>
      <select name="per_page">
        <option value="20" {{if eq .PerPage 20}}selected{{end}}>20</option>
        <option value="50" {{if eq .PerPage 50}}selected{{end}}>50</option>
        <option value="100" {{if eq .PerPage 100}}selected{{end}}>100</option>
      </select>
    </div>
    <div class="form-group"><label>&nbsp;</label><button type="submit" class="btn btn-primary">Filtrar</button></div>
  </div></form>
</div></div>

<div class="card"><div class="card-header"><h3>Historial de Violaciones ({{.Total}} total)</h3></div><div class="card-body">
{{if .Violations}}
<div class="table-responsive"><table class="data-table"><thead><tr><th>Fecha</th><th>Servidor</th><th>Jugador</th><th>IP</th><th>Tipo</th><th>Razón</th><th>Perfil</th></tr></thead>
<tbody>{{range .Violations}}<tr>
  <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
  <td>{{.ServerAddr}}</td>
  <td><a href="/player?q={{.PlayerName}}" class="player-link" style="font-weight:600;">{{.PlayerName}}</a></td>
  <td><code>{{.PlayerIP}}</code></td>
  <td><span class="badge badge-{{if eq .Type "file"}}warning{{else if eq .Type "cvar"}}danger{{else}}info{{end}}">{{.Type}}</span></td>
  <td style="font-size:13px;">{{.Reason}}</td>
  <td><a href="/player?q={{.PlayerName}}" class="btn btn-sm">&#128100; Ver</a></td>
</tr>{{end}}</tbody></table></div>
{{if gt .TotalPages 1}}<div class="pagination">
  {{if gt .Page 1}}<a href="?page={{sub .Page 1}}&per_page={{.PerPage}}&name={{.PlayerName}}&player={{.PlayerIP}}&server={{.ServerAddr}}&type={{.Type}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Anterior</a>{{end}}
  <span class="page-info">Página {{.Page}} de {{.TotalPages}}</span>
  {{if lt .Page .TotalPages}}<a href="?page={{add .Page 1}}&per_page={{.PerPage}}&name={{.PlayerName}}&player={{.PlayerIP}}&server={{.ServerAddr}}&type={{.Type}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Siguiente</a>{{end}}
</div>{{end}}
{{else}}<div class="empty-state"><p>No se encontraron violaciones con los filtros aplicados</p></div>{{end}}
</div></div>

</div></div>
<script src="/static/app.js"></script>
</body></html>`,

"servers": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0,maximum-scale=1.0,user-scalable=no">
<title>Dday AC - Servers</title>
<meta name="theme-color" content="#1e293b">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="Dday AC">
<link rel="manifest" href="/manifest.json">
<link rel="icon" type="image/svg+xml" href="/static/icon.svg">
<link rel="apple-touch-icon" href="/static/icon.svg">
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Servers</h2></div>
<div class="content">
<div class="card"><div class="card-header"><h3>Servidores Conectados</h3><button class="btn btn-sm" onclick="location.reload()">&#8635; Actualizar</button></div><div class="card-body">
{{if .Servers}}{{range .Servers}}
<div class="server-card">
  <div class="server-header">
    <div class="server-name">{{.Hostname}}</div>
    <div class="server-version">v{{.Version}}</div>
    <div class="server-addr">{{.RemoteAddr}}</div>
    <div class="server-port">Puerto: {{.Port}}</div>
  </div>
  <div class="server-clients"><h4>Clientes Conectados ({{len .Clients}})</h4>
  {{if .Clients}}
  <div class="table-responsive"><table class="data-table compact"><thead><tr><th>ID</th><th>Nombre</th><th>IP</th><th>Tipo</th><th>Válido</th><th>Fallos</th><th>Acción</th></tr></thead>
  <tbody>{{range $id, $client := .Clients}}<tr>
    <td>{{$client.ClientID}}</td>
    <td><a href="/player?q={{$client.Name}}" class="player-link" style="font-weight:600;">{{$client.Name}}</a></td>
    <td><code>{{$client.IP}}</code></td>
    <td>{{$client.ClientTypeString}}</td>
    <td>{{if $client.Valid}}<span class="badge badge-success">Sí</span>{{else}}<span class="badge badge-danger">No</span>{{end}}</td>
    <td>{{$client.FileFailures}}</td>
    <td><a href="/player?q={{$client.Name}}" class="btn btn-sm">&#128100; Perfil</a></td>
  </tr>{{end}}</tbody></table></div>
  {{else}}<div class="empty-state small"><p>No hay clientes conectados</p></div>{{end}}
  </div>
</div>
{{end}}{{else}}<div class="empty-state"><p>No hay servidores conectados actualmente</p></div>{{end}}
</div></div>
</div></div>
<script src="/static/app.js"></script>
<script>setTimeout(function(){location.reload()},5000)</script>
</body></html>`,

"process-snapshots": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0,maximum-scale=1.0,user-scalable=no">
<title>Dday AC - Process Snapshots</title>
<meta name="theme-color" content="#1e293b">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="Dday AC">
<link rel="manifest" href="/manifest.json">
<link rel="icon" type="image/svg+xml" href="/static/icon.svg">
<link rel="apple-touch-icon" href="/static/icon.svg">
<link rel="stylesheet" href="/static/style.css">
<style>
.violation-preview { font-size: 11px; color: #f87171; margin-top: 4px; line-height: 1.3; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.violation-count { font-weight: 600; }
</style>
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Process Snapshots</h2></div>
<div class="content">

<div class="card"><div class="card-header"><h3>Filtros de Búsqueda</h3></div><div class="card-body">
  <form method="GET" action="/process-snapshots" class="filter-form"><div class="form-row">
    <div class="form-group"><label>Nombre Jugador (parcial/exacto)</label><input type="text" name="name" value="{{.PlayerName}}" placeholder="Buscar nombre..."></div>
    <div class="form-group"><label>IP Jugador</label><input type="text" name="player" value="{{.PlayerIP}}" placeholder="Filtrar por IP..."></div>
    <div class="form-group"><label>Servidor</label>
      <select name="server">
        <option value="">Todos los servidores</option>
        {{range .Servers}}
        <option value="{{.}}" {{if eq $.ServerAddr .}}selected{{end}}>{{.}}</option>
        {{end}}
      </select>
    </div>
    <div class="form-group"><label>Desde</label><input type="date" name="from" value="{{.DateFrom}}"></div>
    <div class="form-group"><label>Hasta</label><input type="date" name="to" value="{{.DateTo}}"></div>
    <div class="form-group"><label>Por página</label>
      <select name="per_page">
        <option value="20" {{if eq .PerPage 20}}selected{{end}}>20</option>
        <option value="50" {{if eq .PerPage 50}}selected{{end}}>50</option>
        <option value="100" {{if eq .PerPage 100}}selected{{end}}>100</option>
      </select>
    </div>
    <div class="form-group"><label>&nbsp;</label><button type="submit" class="btn btn-primary">Filtrar</button></div>
  </div></form>
</div></div>

<div class="card"><div class="card-header"><h3>Snapshots Registrados ({{.Total}} total)</h3></div><div class="card-body">
{{if .Snapshots}}
<div class="table-responsive"><table class="data-table">
  <thead><tr><th>ID</th><th>Servidor</th><th>Jugador</th><th>IP</th><th>Procesos</th><th>Módulos</th><th>Violaciones</th><th>Fecha</th><th>Acción</th></tr></thead>
  <tbody>{{range .Snapshots}}
  <tr>
    <td><strong>#{{.ID}}</strong></td>
    <td>{{.ServerAddr}}</td>
    <td><a href="/player?q={{.PlayerName}}" class="player-link" style="font-weight:600;">{{.PlayerName}}</a></td>
    <td><code>{{.PlayerIP}}</code></td>
    <td>{{.NumProcesses}}</td>
    <td>{{.NumModules}}</td>
    <td>{{if .Violations}}<span class="badge badge-danger violation-count">{{violationsCount .Violations}} violaciones</span><div class="violation-preview" title="{{.Violations}}">{{violationsPreview .Violations}}</div>{{else}}<span class="badge badge-success">Clean</span>{{end}}</td>
    <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
    <td><a href="/process-snapshots/{{.ID}}" class="btn btn-sm">&#128065; Inspeccionar</a></td>
  </tr>
  {{end}}</tbody>
</table></div>
{{if gt .TotalPages 1}}
<div class="pagination">
  {{if gt .Page 1}}<a href="?page={{sub .Page 1}}&per_page={{.PerPage}}&name={{.PlayerName}}&player={{.PlayerIP}}&server={{.ServerAddr}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Anterior</a>{{end}}
  <span class="page-info">Página {{.Page}} de {{.TotalPages}}</span>
  {{if .HasNext}}<a href="?page={{add .Page 1}}&per_page={{.PerPage}}&name={{.PlayerName}}&player={{.PlayerIP}}&server={{.ServerAddr}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Siguiente</a>{{end}}
</div>
{{end}}
{{else}}<div class="empty-state"><p>No se encontraron process snapshots con los filtros aplicados</p></div>{{end}}
</div></div>

</div></div>
<script src="/static/app.js"></script>
</body></html>`,

"process-snapshot-detail": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0,maximum-scale=1.0,user-scalable=no">
<title>Dday AC - Process Snapshot Detail</title>
<meta name="theme-color" content="#1e293b">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="Dday AC">
<link rel="manifest" href="/manifest.json">
<link rel="icon" type="image/svg+xml" href="/static/icon.svg">
<link rel="apple-touch-icon" href="/static/icon.svg">
<link rel="stylesheet" href="/static/style.css">
<style>
.process-table, .module-table { width: 100%; border-collapse: collapse; margin-top: 10px; }
.process-table th, .module-table th { background: #1a1d29; color: #8b95a5; padding: 10px 12px; text-align: left; font-size: 12px; text-transform: uppercase; }
.process-table td, .module-table td { padding: 8px 12px; border-bottom: 1px solid #2a2d3a; font-size: 13px; }
.process-table tr:hover, .module-table tr:hover { background: #1e2130; }
.row-suspicious { background: rgba(239, 68, 68, 0.1) !important; }
.row-suspicious td { color: #ef4444; font-weight: 500; }
.badge-suspicious { background: #ef4444; color: white; padding: 2px 8px; border-radius: 4px; font-size: 11px; }
.section-title { font-size: 16px; font-weight: 600; margin: 20px 0 10px 0; color: #e2e8f0; }
.snapshot-info { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 16px; margin-bottom: 16px; }
.info-item { background: #1a1d29; padding: 12px 16px; border-radius: 8px; }
.info-label { font-size: 12px; color: #8b95a5; text-transform: uppercase; margin-bottom: 4px; }
.info-value { font-size: 14px; color: #e2e8f0; font-weight: 500; }
.violations-summary { background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); border-radius: 8px; padding: 10px 16px; margin-bottom: 20px; display: flex; align-items: center; gap: 8px; }
.violations-summary .count { color: #ef4444; font-weight: 600; font-size: 14px; }
.violations-summary .label { color: #f87171; font-size: 13px; }
.violations-list { background: #1a1d29; border-radius: 8px; padding: 16px; margin-bottom: 20px; }
.violation-item { padding: 8px 0; border-bottom: 1px solid #2a2d3a; font-size: 13px; color: #e2e8f0; display: flex; align-items: flex-start; gap: 8px; }
.violation-item:last-child { border-bottom: none; }
.violation-icon { color: #ef4444; font-size: 14px; flex-shrink: 0; margin-top: 1px; }
.violation-text { line-height: 1.4; }
</style>
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Process Snapshot #{{.Snapshot.ID}}</h2></div>
<div class="content">
<div class="card"><div class="card-header">
  <h3>Detalle del Snapshot</h3>
  <div style="display:flex; gap:8px;">
    <a href="/player?q={{.Snapshot.PlayerName}}" class="btn btn-sm btn-primary">&#128100; Ver Perfil</a>
    <a href="/process-snapshots" class="btn btn-sm">Volver</a>
  </div>
</div><div class="card-body">
  <div class="snapshot-info">
    <div class="info-item"><div class="info-label">Jugador</div><div class="info-value"><a href="/player?q={{.Snapshot.PlayerName}}" class="player-link">{{.Snapshot.PlayerName}}</a></div></div>
    <div class="info-item"><div class="info-label">IP</div><div class="info-value"><code>{{.Snapshot.PlayerIP}}</code></div></div>
    <div class="info-item"><div class="info-label">Servidor</div><div class="info-value">{{.Snapshot.ServerAddr}}</div></div>
    <div class="info-item"><div class="info-label">Fecha</div><div class="info-value">{{.Snapshot.Timestamp.Format "2006-01-02 15:04:05"}}</div></div>
    <div class="info-item"><div class="info-label">Total Procesos</div><div class="info-value">{{.Snapshot.NumProcesses}}</div></div>
    <div class="info-item"><div class="info-label">Total Módulos</div><div class="info-value">{{.Snapshot.NumModules}}</div></div>
  </div>

  {{if .Snapshot.Violations}}
  <div class="violations-summary">
    <span class="count">&#9888; {{violationsCount .Snapshot.Violations}} violaciones detectadas</span>
  </div>
  <div class="section-title">Violaciones</div>
  <div class="violations-list">
    {{range violationsList .Snapshot.Violations}}
    <div class="violation-item">
      <span class="violation-icon">&#9679;</span>
      <div class="violation-text">{{.}}</div>
    </div>
    {{end}}
  </div>
  {{end}}

  <div class="section-title">Procesos ({{len .Processes}})</div>
  {{if .Processes}}
  <div class="table-responsive"><table class="process-table data-table">
    <thead><tr><th>PID</th><th>PID Padre</th><th>Nombre</th><th>Estado</th><th>Acción</th></tr></thead>
    <tbody>{{range .Processes}}
    <tr{{if .Suspicious}} class="row-suspicious"{{end}}>
      <td>{{.PID}}</td>
      <td>{{.ParentPID}}</td>
      <td>{{.Name}}{{if .Suspicious}} <span class="badge-suspicious">{{.MatchPattern}}</span>{{end}}</td>
      <td>{{if .Suspicious}}<span style="color:#ef4444;font-weight:600">SOSPECHOSO</span>{{else}}<span style="color:#22c55e;font-weight:600">LIMPIO</span>{{end}}</td>
      <td>
        <form method="POST" action="/blacklist" style="display:inline" onsubmit="return confirm('¿Agregar \'{{.Name}}\' a la Blacklist de procesos?')">
          <input type="hidden" name="action" value="add">
          <input type="hidden" name="type" value="process">
          <input type="hidden" name="pattern" value="{{.Name}}">
          <input type="hidden" name="added_by" value="snapshot_inspect">
          <button type="submit" class="btn btn-sm btn-outline-danger" title="Agregar proceso a la blacklist">&#128683; Blacklist</button>
        </form>
      </td>
    </tr>
    {{end}}</tbody>
  </table></div>
  {{else}}<p style="color:#8b95a5">No hay datos de procesos</p>{{end}}

  <div class="section-title">Módulos ({{len .Modules}})</div>
  {{if .Modules}}
  <div class="table-responsive"><table class="module-table data-table">
    <thead><tr><th>Nombre</th><th>Ruta</th><th>SHA1</th><th>Estado</th><th>Acción</th></tr></thead>
    <tbody>{{range .Modules}}
    <tr{{if .Suspicious}} class="row-suspicious"{{end}}>
      <td>{{.Name}}{{if .Suspicious}} <span class="badge-suspicious">{{.MatchPattern}}</span>{{end}}</td>
      <td style="font-size:11px;color:#8b95a5">{{.Path}}</td>
      <td style="font-family:monospace;font-size:11px">{{.SHA1}}</td>
      <td>{{if .Suspicious}}<span style="color:#ef4444;font-weight:600">SOSPECHOSO</span>{{else}}<span style="color:#22c55e;font-weight:600">LIMPIO</span>{{end}}</td>
      <td>
        <div style="display:flex; gap:6px; flex-wrap:wrap;">
          <form method="POST" action="/blacklist" style="display:inline" onsubmit="return confirm('¿Agregar módulo \'{{.Name}}\' a la Blacklist?')">
            <input type="hidden" name="action" value="add">
            <input type="hidden" name="type" value="module">
            <input type="hidden" name="pattern" value="{{.Name}}">
            <input type="hidden" name="added_by" value="snapshot_inspect">
            <button type="submit" class="btn btn-sm btn-outline-danger" title="Bloquear por nombre de DLL">&#128683; Nombre</button>
          </form>
          {{if .SHA1}}
          <form method="POST" action="/blacklist" style="display:inline" onsubmit="return confirm('¿Agregar SHA1 \'{{.SHA1}}\' a la Blacklist?')">
            <input type="hidden" name="action" value="add">
            <input type="hidden" name="type" value="sha1">
            <input type="hidden" name="pattern" value="{{.SHA1}}">
            <input type="hidden" name="added_by" value="snapshot_inspect">
            <button type="submit" class="btn btn-sm btn-outline-danger" title="Bloquear por hash SHA1 exacto (independiente del nombre)"># SHA1</button>
          </form>
          {{end}}
        </div>
      </td>
    </tr>
    {{end}}</tbody>
  </table></div>
  {{else}}<p style="color:#8b95a5">No hay datos de módulos</p>{{end}}
</div></div>
</div></div>
<script src="/static/app.js"></script>
</body></html>`,

"blacklist": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0,maximum-scale=1.0,user-scalable=no">
<title>Dday AC - Blacklist</title>
<meta name="theme-color" content="#1e293b">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="Dday AC">
<link rel="manifest" href="/manifest.json">
<link rel="icon" type="image/svg+xml" href="/static/icon.svg">
<link rel="apple-touch-icon" href="/static/icon.svg">
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Blacklist de Procesos, Módulos y Hashes SHA1</h2></div>
<div class="content">

{{if eq .Msg "added"}}
<div class="alert alert-success"><span>&#10004;</span> Patrón agregado exitosamente a la blacklist.</div>
{{else if eq .Msg "deleted"}}
<div class="alert alert-success"><span>&#10004;</span> Patrón eliminado correctamente de la blacklist.</div>
{{else if eq .Msg "toggled"}}
<div class="alert alert-info"><span>&#9432;</span> Estado del patrón actualizado correctamente.</div>
{{else if eq .Error "empty_pattern"}}
<div class="alert alert-danger"><span>&#9888;</span> El patrón no puede estar vacío.</div>
{{else if eq .Error "add_failed"}}
<div class="alert alert-danger"><span>&#9888;</span> Error al agregar el patrón. Verifique si ya existe en la lista.</div>
{{else if eq .Error "delete_failed"}}
<div class="alert alert-danger"><span>&#9888;</span> Error al eliminar el patrón de la base de datos.</div>
{{else if eq .Error "toggle_failed"}}
<div class="alert alert-danger"><span>&#9888;</span> Error al cambiar el estado del patrón.</div>
{{end}}

<div class="stats-grid">
  <div class="stat-card"><div class="stat-icon red">&#9888;</div><div class="stat-info"><div class="stat-value">{{.ProcessCount}}</div><div class="stat-label">Procesos (.exe)</div></div></div>
  <div class="stat-card"><div class="stat-icon orange">&#128737;</div><div class="stat-info"><div class="stat-value">{{.ModuleCount}}</div><div class="stat-label">Módulos (.dll)</div></div></div>
  <div class="stat-card"><div class="stat-icon purple">&#128273;</div><div class="stat-info"><div class="stat-value">{{.SHA1Count}}</div><div class="stat-label">Hashes SHA1</div></div></div>
  <div class="stat-card"><div class="stat-icon blue">&#128196;</div><div class="stat-info"><div class="stat-value">{{.TotalEntries}}</div><div class="stat-label">Total Registrados</div></div></div>
</div>

<div class="card"><div class="card-header"><h3>Agregar Nuevo Patrón o Hash a la Blacklist</h3></div><div class="card-body">
  <form method="POST" action="/blacklist" class="filter-form">
    <input type="hidden" name="action" value="add">
    <div class="form-row">
      <div class="form-group" style="max-width: 180px;"><label>Tipo de Elemento</label>
        <select name="type" required>
          <option value="process">Proceso (.exe)</option>
          <option value="module">Módulo (.dll)</option>
          <option value="sha1">Hash SHA1 (40 hex)</option>
        </select>
      </div>
      <div class="form-group" style="flex: 2;"><label>Patrón o Hash a Bloquear</label>
        <input type="text" name="pattern" required placeholder="ej: cheatengine, aimware, xenos.dll, o 40 caracteres hex de SHA1">
      </div>
      <div class="form-group" style="max-width: 180px;"><label>Registrado por</label>
        <input type="text" name="added_by" value="admin" placeholder="admin">
      </div>
      <div class="form-group" style="flex: 0;"><label>&nbsp;</label>
        <button type="submit" class="btn btn-primary" style="white-space: nowrap;">&#43; Agregar Patrón</button>
      </div>
    </div>
  </form>
</div></div>

<div class="card">
  <div class="card-header" style="display:flex; justify-content:space-between; align-items:center; flex-wrap:wrap; gap:12px;">
    <h3>Patrones Configurados (<span id="visible-count">{{.TotalEntries}}</span> / {{.TotalEntries}})</h3>
    <div style="display:flex; gap:10px; align-items:center; flex-wrap:wrap;">
      <div class="search-box">
        <span class="search-icon">&#128269;</span>
        <input type="text" id="bl-search" placeholder="Filtrar patrones..." onkeyup="filterBlacklistTable()">
      </div>
      <select id="bl-filter-type" onchange="filterBlacklistTable()" style="padding:7px 10px; background:var(--bg-primary); border:1px solid var(--border); border-radius:6px; color:var(--text-primary); font-size:13px;">
        <option value="all">Todos los tipos</option>
        <option value="process">Solo Procesos</option>
        <option value="module">Solo Módulos</option>
        <option value="sha1">Solo Hashes SHA1</option>
      </select>
      <select id="bl-filter-status" onchange="filterBlacklistTable()" style="padding:7px 10px; background:var(--bg-primary); border:1px solid var(--border); border-radius:6px; color:var(--text-primary); font-size:13px;">
        <option value="all">Todos los estados</option>
        <option value="active">Solo Activos</option>
        <option value="inactive">Solo Inactivos</option>
      </select>
    </div>
  </div>
  <div class="card-body">
{{if .Entries}}
<div class="table-responsive"><table class="data-table" id="blacklist-table">
  <thead><tr>
    <th style="width: 60px;">ID</th>
    <th style="width: 110px;">Tipo</th>
    <th>Patrón / Hash de Coincidencia</th>
    <th style="width: 110px;">Origen</th>
    <th style="width: 110px;">Estado</th>
    <th style="width: 130px;">Agregado por</th>
    <th style="width: 120px; text-align: right;">Acciones</th>
  </tr></thead>
  <tbody>{{range .Entries}}
  <tr{{if not .Enabled}} class="row-disabled"{{end}} data-type="{{.Type}}" data-status="{{if .Enabled}}active{{else}}inactive{{end}}" data-pattern="{{.Pattern}}">
    <td>{{.ID}}</td>
    <td><span class="badge {{if eq .Type "process"}}badge-danger{{else if eq .Type "sha1"}}badge-info{{else}}badge-warning{{end}}">{{if eq .Type "process"}}Proceso{{else if eq .Type "sha1"}}SHA1 Hash{{else}}Módulo{{end}}</span></td>
    <td><code style="font-weight:600; font-size:13px;">{{.Pattern}}</code></td>
    <td>{{if eq .Source "hardcoded"}}<span class="badge-hardcoded">Sistema</span>{{else}}<span class="badge-user">Usuario</span>{{end}}</td>
    <td>
      <form method="POST" action="/blacklist" style="display:inline">
        <input type="hidden" name="action" value="toggle">
        <input type="hidden" name="id" value="{{.ID}}">
        <button type="submit" class="toggle-btn {{if .Enabled}}toggle-active{{else}}toggle-inactive{{end}}" title="Haga clic para activar o desactivar">
          {{if .Enabled}}Activo{{else}}Inactivo{{end}}
        </button>
      </form>
    </td>
    <td><span style="color:var(--text-secondary);">{{.AddedBy}}</span></td>
    <td style="text-align: right;">
      <form method="POST" action="/blacklist" style="display:inline" onsubmit="return confirm('¿Está seguro de eliminar el patrón \'{{.Pattern}}\' de la blacklist?')">
        <input type="hidden" name="action" value="delete">
        <input type="hidden" name="id" value="{{.ID}}">
        <button type="submit" class="btn btn-sm btn-danger" title="Eliminar este patrón de la blacklist">
          &#128465; Eliminar
        </button>
      </form>
    </td>
  </tr>
  {{end}}</tbody>
</table></div>
<div id="no-filter-results" style="display:none; padding:24px; text-align:center; color:var(--text-secondary);">
  <p>No se encontraron patrones que coincidan con el filtro de búsqueda.</p>
</div>
{{else}}<div class="empty-state"><p>No hay patrones en la blacklist</p></div>{{end}}
</div></div>
</div></div>
<script src="/static/app.js"></script>
<script>
function filterBlacklistTable() {
  var searchInput = document.getElementById('bl-search').value.toLowerCase().trim();
  var typeFilter = document.getElementById('bl-filter-type').value;
  var statusFilter = document.getElementById('bl-filter-status').value;
  var table = document.getElementById('blacklist-table');
  if (!table) return;

  var rows = table.querySelectorAll('tbody tr');
  var visibleCount = 0;

  rows.forEach(function(row) {
    var type = row.getAttribute('data-type');
    var status = row.getAttribute('data-status');
    var pattern = (row.getAttribute('data-pattern') || '').toLowerCase();

    var matchSearch = !searchInput || pattern.indexOf(searchInput) !== -1;
    var matchType = typeFilter === 'all' || type === typeFilter;
    var matchStatus = statusFilter === 'all' || status === statusFilter;

    if (matchSearch && matchType && matchStatus) {
      row.style.display = '';
      visibleCount++;
    } else {
      row.style.display = 'none';
    }
  });

  var countSpan = document.getElementById('visible-count');
  if (countSpan) countSpan.textContent = visibleCount;

  var noResults = document.getElementById('no-filter-results');
  if (noResults) {
    noResults.style.display = (visibleCount === 0 && rows.length > 0) ? 'block' : 'none';
  }
}
</script>
</body></html>`,

"settings": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0,maximum-scale=1.0,user-scalable=no">
<title>Dday AC - Configuración</title>
<meta name="theme-color" content="#1e293b">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="Dday AC">
<link rel="manifest" href="/manifest.json">
<link rel="icon" type="image/svg+xml" href="/static/icon.svg">
<link rel="apple-touch-icon" href="/static/icon.svg">
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Configuración y Mantenimiento</h2></div>
<div class="content">

{{if eq .Msg "password_changed"}}
<div class="alert alert-success"><span>&#10004;</span> La contraseña de administrador se ha cambiado exitosamente.</div>
{{else if eq .Msg "purged"}}
<div class="alert alert-success"><span>&#10004;</span> Capturas antiguas purgadas exitosamente del disco y base de datos.</div>
{{else if eq .Error "password_mismatch"}}
<div class="alert alert-danger"><span>&#9888;</span> La nueva contraseña y la confirmación no coinciden.</div>
{{else if eq .Error "invalid_old_password"}}
<div class="alert alert-danger"><span>&#9888;</span> La contraseña actual ingresada es incorrecta.</div>
{{else if eq .Error "empty_password"}}
<div class="alert alert-danger"><span>&#9888;</span> La nueva contraseña no puede estar vacía.</div>
{{else if eq .Error "change_failed"}}
<div class="alert alert-danger"><span>&#9888;</span> Error al actualizar la contraseña en la base de datos.</div>
{{else if eq .Error "purge_failed"}}
<div class="alert alert-danger"><span>&#9888;</span> Error durante la purga de capturas.</div>
{{end}}

<div class="grid-2">
  <div class="card">
    <div class="card-header"><h3>&#128274; Seguridad - Cambiar Contraseña de Admin</h3></div>
    <div class="card-body">
      <form method="POST" action="/change-password">
        <div class="form-group" style="margin-bottom:14px;">
          <label>Contraseña Actual</label>
          <input type="password" name="old_password" required placeholder="Contraseña actual" style="width:100%; padding:9px 12px; border-radius:6px; background:var(--bg-primary); border:1px solid var(--border); color:var(--text-primary);">
        </div>
        <div class="form-group" style="margin-bottom:14px;">
          <label>Nueva Contraseña</label>
          <input type="password" name="new_password" required placeholder="Nueva contraseña segura" style="width:100%; padding:9px 12px; border-radius:6px; background:var(--bg-primary); border:1px solid var(--border); color:var(--text-primary);">
        </div>
        <div class="form-group" style="margin-bottom:20px;">
          <label>Confirmar Nueva Contraseña</label>
          <input type="password" name="confirm_password" required placeholder="Repita la nueva contraseña" style="width:100%; padding:9px 12px; border-radius:6px; background:var(--bg-primary); border:1px solid var(--border); color:var(--text-primary);">
        </div>
        <button type="submit" class="btn btn-primary">&#128190; Actualizar Contraseña</button>
      </form>
    </div>
  </div>

  <div class="card">
    <div class="card-header"><h3>&#128450; Mantenimiento de Almacenamiento</h3></div>
    <div class="card-body">
      <div style="margin-bottom:16px;">
        <div class="info-row"><span>Total capturas en disco:</span><strong>{{index .Stats "total_screenshots"}} archivos</strong></div>
        <div class="info-row"><span>Espacio en disco ocupado:</span><strong id="total-size">{{index .Stats "total_size"}} bytes</strong></div>
      </div>
      <hr style="border:0; border-top:1px solid var(--border); margin:16px 0;">
      <h4 style="margin:0 0 8px 0; font-size:14px; color:var(--text-primary);">Purga Automática de Screenshots Antiguos</h4>
      <p style="font-size:12px; color:var(--text-secondary); margin-bottom:14px;">
        Elimina las imágenes y registros de screenshots más antiguos que el período seleccionado para liberar espacio en disco.
      </p>
      <form method="POST" action="/maintenance/purge-screenshots" onsubmit="return confirm('¿Está seguro de eliminar permanentemente los screenshots anteriores al plazo seleccionado?')">
        <div class="form-group" style="margin-bottom:14px;">
          <label>Eliminar screenshots con más de:</label>
          <select name="days" style="width:100%; padding:9px 12px; border-radius:6px; background:var(--bg-primary); border:1px solid var(--border); color:var(--text-primary);">
            <option value="30">30 días (1 mes)</option>
            <option value="60">60 días (2 meses)</option>
            <option value="90">90 días (3 meses)</option>
            <option value="180">180 días (6 meses)</option>
            <option value="365">365 días (1 año)</option>
          </select>
        </div>
        <button type="submit" class="btn btn-danger">&#128465; Purgar Archivos Antiguos</button>
      </form>
    </div>
  </div>
</div>

<div class="card" style="margin-top:24px;">
  <div class="card-header"><h3>🔔 Notificaciones Push en Vivo (Navegador y PWA)</h3></div>
  <div class="card-body">
    <p style="font-size:13px; color:var(--text-secondary); margin-bottom:14px;">
      Recibe alertas instantáneas en tu celular o escritorio cada vez que se detecte una violación de un jugador en tiempo real.
    </p>
    <div class="info-row" style="display:flex; justify-content:space-between; align-items:center; padding:8px 0; border-bottom:1px solid var(--border);">
      <span>Estado de suscripción en este dispositivo:</span>
      <span id="push-status-text">Comprobando...</span>
    </div>
    <div style="display:flex; gap:12px; margin-top:16px; flex-wrap:wrap;">
      <button type="button" class="push-toggle-btn" onclick="togglePushSubscription()" style="padding:8px 16px; font-size:14px;">
        🔔 Activar / Desactivar Notificaciones Push
      </button>
      <button type="button" class="btn btn-outline" onclick="sendTestPushNotification()" style="padding:8px 16px; font-size:14px;">
        🧪 Enviar Notificación de Prueba
      </button>
    </div>
  </div>
</div>

</div></div>
<script src="/static/app.js"></script>
</body></html>`,

"sidebar": `<button class="hamburger" onclick="toggleSidebar()">&#9776;</button>
<div class="sidebar-overlay" onclick="toggleSidebar()"></div>
<div class="sidebar">
  <div class="sidebar-header">
    <a href="/" class="logo-link">
      <img src="/static/logo.svg" alt="Dday AC Logo" class="logo-img">
    </a>
  </div>
  <nav class="sidebar-nav">
    <a href="/" class="nav-item {{if eq .CurrentPage "dashboard"}}active{{end}}"><span class="nav-icon">&#9632;</span> Dashboard</a>
    <a href="/screenshots" class="nav-item {{if eq .CurrentPage "screenshots"}}active{{end}}"><span class="nav-icon">&#128247;</span> Screenshots</a>
    <a href="/violations" class="nav-item {{if eq .CurrentPage "violations"}}active{{end}}"><span class="nav-icon">&#9888;</span> Violations</a>
    <a href="/process-snapshots" class="nav-item {{if eq .CurrentPage "process-snapshots"}}active{{end}}"><span class="nav-icon">&#128737;</span> Processes</a>
    <a href="/blacklist" class="nav-item {{if eq .CurrentPage "blacklist"}}active{{end}}"><span class="nav-icon">&#128683;</span> Blacklist</a>
    <a href="/servers" class="nav-item {{if eq .CurrentPage "servers"}}active{{end}}"><span class="nav-icon">&#127760;</span> Servers</a>
    <a href="/settings" class="nav-item {{if eq .CurrentPage "settings"}}active{{end}}"><span class="nav-icon">&#9881;</span> Configuración</a>
  </nav>
  <div class="sidebar-footer">
    <button type="button" onclick="triggerPWAInstall()" class="nav-item" style="width:100%;text-align:left;background:none;border:none;cursor:pointer;color:var(--text-secondary);font-size:13px;padding:8px 20px;display:flex;align-items:center;gap:12px;"><span class="nav-icon">&#128241;</span> Instalar App</button>
    <button type="button" onclick="forceResetPWA()" class="nav-item" style="width:100%;text-align:left;background:none;border:none;cursor:pointer;color:var(--text-secondary);font-size:13px;padding:8px 20px;display:flex;align-items:center;gap:12px;"><span class="nav-icon">&#128260;</span> Actualizar PWA</button>
    <a href="/logout" class="nav-item logout"><span class="nav-icon">&#10140;</span> Salir</a>
  </div>
</div>

<!-- Mobile Bottom Navigation Bar -->
<nav class="bottom-nav">
  <a href="/" class="bottom-nav-item {{if eq .CurrentPage "dashboard"}}active{{end}}"><span class="bnav-icon">&#127968;</span><span>Inicio</span></a>
  <a href="/screenshots" class="bottom-nav-item {{if eq .CurrentPage "screenshots"}}active{{end}}"><span class="bnav-icon">&#128247;</span><span>Capturas</span></a>
  <a href="/violations" class="bottom-nav-item {{if eq .CurrentPage "violations"}}active{{end}}"><span class="bnav-icon">&#9888;</span><span>Alertas</span></a>
  <a href="/process-snapshots" class="bottom-nav-item {{if eq .CurrentPage "process-snapshots"}}active{{end}}"><span class="bnav-icon">&#128737;</span><span>Procesos</span></a>
  <a href="#" onclick="toggleSidebar();return false;" class="bottom-nav-item"><span class="bnav-icon">&#9776;</span><span>Más</span></a>
</nav>

<!-- PWA Install Banner -->
<div id="pwa-install-banner" class="pwa-banner">
  <div class="pwa-banner-content">
    <div class="pwa-banner-icon" style="background:transparent;padding:0;"><img src="/static/icon.svg" style="width:40px;height:40px;border-radius:8px;" alt="Dday AC"></div>
    <div class="pwa-banner-text">
      <h4>Instalar Dday AC</h4>
      <p>Administra y modera desde tu celular</p>
    </div>
  </div>
  <div class="pwa-banner-actions">
    <button type="button" class="btn btn-sm btn-primary" onclick="triggerPWAInstall()">Instalar</button>
    <button type="button" class="pwa-close-btn" onclick="dismissInstallBanner()" title="Cerrar">&times;</button>
  </div>
</div>`,
}
