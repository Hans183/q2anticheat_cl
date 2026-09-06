package web

var templates = map[string]string{
"login": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Anticheat - Login</title>
<link rel="stylesheet" href="/static/style.css">
</head><body class="login-body">
<div class="login-container">
  <div class="login-header">
    <div class="login-icon">&#128737;</div>
    <h1>Q2PRO Anticheat</h1>
    <p>Dashboard de Administracion</p>
  </div>
  {{if .Error}}<div class="alert alert-error">{{.Error}}</div>{{end}}
  <form method="POST" action="/login">
    <div class="form-group"><label>Usuario</label>
      <input type="text" name="username" required autofocus placeholder="admin"></div>
    <div class="form-group"><label>Contrasena</label>
      <input type="password" name="password" required placeholder="&bull;&bull;&bull;&bull;&bull;&bull;"></div>
    <button type="submit" class="btn btn-primary btn-block">Entrar</button>
  </form>
</div></body></html>`,

"dashboard": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Q2PRO Anticheat - Dashboard</title>
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Dashboard</h2></div>
<div class="content">
<div class="stats-grid">
  <div class="stat-card"><div class="stat-icon blue">&#127760;</div><div class="stat-info"><div class="stat-value">{{.ServerCount}}</div><div class="stat-label">Servidores Conectados</div></div></div>
  <div class="stat-card"><div class="stat-icon green">&#128100;</div><div class="stat-info"><div class="stat-value">{{.ClientCount}}</div><div class="stat-label">Jugadores Activos</div></div></div>
  <div class="stat-card"><div class="stat-icon purple">&#128247;</div><div class="stat-info"><div class="stat-value">{{index .Stats "total_screenshots"}}</div><div class="stat-label">Total Screenshots</div></div></div>
  <div class="stat-card"><div class="stat-icon orange">&#9888;</div><div class="stat-info"><div class="stat-value">{{index .Stats "total_violations"}}</div><div class="stat-label">Total Violations</div></div></div>
  <div class="stat-card"><div class="stat-icon purple">&#128737;</div><div class="stat-info"><div class="stat-value">{{index .Stats "total_process_snapshots"}}</div><div class="stat-label">Process Snapshots</div></div></div>
</div>
<div class="grid-2">
  <div class="card"><div class="card-header"><h3>Resumen</h3></div><div class="card-body">
    <div class="info-row"><span>Screenshots sin revisar</span><span class="badge badge-warning">{{index .Stats "unreviewed_screenshots"}}</span></div>
    <div class="info-row"><span>Violations hoy</span><span class="badge badge-danger">{{index .Stats "today_violations"}}</span></div>
    <div class="info-row"><span>Espacio utilizado</span><span class="badge badge-info" id="total-size">{{index .Stats "total_size"}} bytes</span></div>
  </div></div>
  <div class="card"><div class="card-header"><h3>Accesos Rapidos</h3></div><div class="card-body">
    <a href="/screenshots?unreviewed=1" class="quick-link"><span class="ql-icon">&#128247;</span><span>Screenshots sin revisar</span></a>
    <a href="/violations?type=file" class="quick-link"><span class="ql-icon">&#128196;</span><span>Violaciones de archivos</span></a>
    <a href="/violations?type=cvar" class="quick-link"><span class="ql-icon">&#128260;</span><span>Violaciones de cvars</span></a>
    <a href="/process-snapshots" class="quick-link"><span class="ql-icon">&#128737;</span><span>Process Snapshots</span></a>
    <a href="/servers" class="quick-link"><span class="ql-icon">&#127760;</span><span>Estado de servidores</span></a>
  </div></div>
</div>
</div></div>
<script src="/static/app.js"></script>
</body></html>`,

"screenshots": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Q2PRO Anticheat - Screenshots</title>
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Screenshots</h2></div>
<div class="content">
<div class="card"><div class="card-header"><h3>Filtros</h3></div><div class="card-body">
  <form method="GET" action="/screenshots" class="filter-form"><div class="form-row">
    <div class="form-group"><label>Nombre Jugador</label><input type="text" name="name" value="{{.PlayerName}}" placeholder="Buscar nombre..."></div>
    <div class="form-group"><label>IP Jugador</label><input type="text" name="player" value="{{.PlayerIP}}" placeholder="Filtrar por IP..."></div>
    <div class="form-group"><label>Desde</label><input type="date" name="from" value="{{.DateFrom}}"></div>
    <div class="form-group"><label>Hasta</label><input type="date" name="to" value="{{.DateTo}}"></div>
    <div class="form-group"><label>&nbsp;</label><label class="checkbox-label"><input type="checkbox" name="unreviewed" value="1" {{if .Unreviewed}}checked{{end}}> Solo sin revisar</label></div>
    <div class="form-group"><label>&nbsp;</label><button type="submit" class="btn btn-primary">Filtrar</button></div>
  </div></form>
</div></div>
<div class="card"><div class="card-header"><h3>Screenshots ({{.Total}} total)</h3></div><div class="card-body">
{{if .Screenshots}}
<div class="screenshot-grid">{{range .Screenshots}}
  <div class="screenshot-card {{if .Reviewed}}reviewed{{end}}">
    <div class="screenshot-img" onclick="openLightbox('/screenshots/image/{{.ID}}','{{.PlayerName}}','{{.PlayerIP}}','{{.Timestamp.Format "2006-01-02 15:04:05"}}','{{.ServerAddr}}')">
      <img src="/screenshots/image/{{.ID}}" alt="Screenshot" loading="lazy">
      {{if not .Reviewed}}<span class="badge badge-new">NUEVO</span>{{end}}
    </div>
    <div class="screenshot-info">
      <div class="ss-name">{{.PlayerName}}</div>
      <div class="ss-ip">{{.PlayerIP}}</div>
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
{{if gt .TotalPages 1}}<div class="pagination">
  {{if gt .Page 1}}<a href="?page={{sub .Page 1}}&name={{.PlayerName}}&player={{.PlayerIP}}&from={{.DateFrom}}&to={{.DateTo}}{{if .Unreviewed}}&unreviewed=1{{end}}" class="btn btn-sm">Anterior</a>{{end}}
  <span class="page-info">Pagina {{.Page}} de {{.TotalPages}}</span>
  {{if lt .Page .TotalPages}}<a href="?page={{add .Page 1}}&name={{.PlayerName}}&player={{.PlayerIP}}&from={{.DateFrom}}&to={{.DateTo}}{{if .Unreviewed}}&unreviewed=1{{end}}" class="btn btn-sm">Siguiente</a>{{end}}
</div>{{end}}
{{else}}<div class="empty-state"><p>No se encontraron screenshots</p></div>{{end}}
</div></div>
<div id="lightbox" class="lightbox" onclick="closeLightbox()">
  <div class="lightbox-content" onclick="event.stopPropagation()">
    <button class="lightbox-close" onclick="closeLightbox()">&times;</button>
    <img id="lightbox-img" src="" alt="Screenshot">
    <div id="lightbox-info" class="lightbox-info"></div>
  </div>
</div>
</div></div>
<script src="/static/app.js"></script>
</body></html>`,

"violations": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Q2PRO Anticheat - Violations</title>
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Violations</h2></div>
<div class="content">
<div class="card"><div class="card-header"><h3>Filtros</h3></div><div class="card-body">
  <form method="GET" action="/violations" class="filter-form"><div class="form-row">
    <div class="form-group"><label>Nombre Jugador</label><input type="text" name="name" value="{{.PlayerName}}" placeholder="Buscar nombre..."></div>
    <div class="form-group"><label>IP Jugador</label><input type="text" name="player" value="{{.PlayerIP}}" placeholder="Filtrar por IP..."></div>
    <div class="form-group"><label>Tipo</label><select name="type">
      <option value="">Todos</option>
      <option value="file" {{if eq .Type "file"}}selected{{end}}>Archivos</option>
      <option value="cvar" {{if eq .Type "cvar"}}selected{{end}}>Cvars</option>
    </select></div>
    <div class="form-group"><label>Desde</label><input type="date" name="from" value="{{.DateFrom}}"></div>
    <div class="form-group"><label>Hasta</label><input type="date" name="to" value="{{.DateTo}}"></div>
    <div class="form-group"><label>&nbsp;</label><button type="submit" class="btn btn-primary">Filtrar</button></div>
  </div></form>
</div></div>
<div class="card"><div class="card-header"><h3>Historial de Violaciones ({{.Total}} total)</h3></div><div class="card-body">
{{if .Violations}}
<div class="table-responsive"><table class="data-table"><thead><tr><th>Fecha</th><th>Servidor</th><th>Jugador</th><th>IP</th><th>Tipo</th><th>Razon</th></tr></thead>
<tbody>{{range .Violations}}<tr>
  <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
  <td>{{.ServerAddr}}</td><td>{{.PlayerName}}</td><td><code>{{.PlayerIP}}</code></td>
  <td><span class="badge badge-{{if eq .Type "file"}}warning{{else if eq .Type "cvar"}}danger{{else}}info{{end}}">{{.Type}}</span></td>
  <td>{{.Reason}}</td>
</tr>{{end}}</tbody></table></div>
{{if gt .TotalPages 1}}<div class="pagination">
  {{if gt .Page 1}}<a href="?page={{sub .Page 1}}&name={{.PlayerName}}&player={{.PlayerIP}}&type={{.Type}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Anterior</a>{{end}}
  <span class="page-info">Pagina {{.Page}} de {{.TotalPages}}</span>
  {{if lt .Page .TotalPages}}<a href="?page={{add .Page 1}}&name={{.PlayerName}}&player={{.PlayerIP}}&type={{.Type}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Siguiente</a>{{end}}
</div>{{end}}
{{else}}<div class="empty-state"><p>No se encontraron violaciones</p></div>{{end}}
</div></div>
</div></div>
<script src="/static/app.js"></script>
</body></html>`,

"servers": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Q2PRO Anticheat - Servers</title>
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Servers</h2></div>
<div class="content">
<div class="card"><div class="card-header"><h3>Servidores Conectados</h3><button class="btn btn-sm" onclick="location.reload()">Actualizar</button></div><div class="card-body">
{{if .Servers}}{{range .Servers}}
<div class="server-card">
  <div class="server-header">
    <div class="server-name">{{.Hostname}}</div>
    <div class="server-version">v{{.Version}}</div>
    <div class="server-addr">{{.RemoteAddr}}</div>
    <div class="server-port">Puerto: {{.Port}}</div>
  </div>
  <div class="server-clients"><h4>Clientes ({{len .Clients}})</h4>
  {{if .Clients}}
  <div class="table-responsive"><table class="data-table compact"><thead><tr><th>ID</th><th>Nombre</th><th>IP</th><th>Tipo</th><th>Valido</th><th>Fallos</th></tr></thead>
  <tbody>{{range $id, $client := .Clients}}<tr>
    <td>{{$client.ClientID}}</td><td>{{$client.Name}}</td><td><code>{{$client.IP}}</code></td>
    <td>{{$client.ClientTypeString}}</td>
    <td>{{if $client.Valid}}<span class="badge badge-success">Si</span>{{else}}<span class="badge badge-danger">No</span>{{end}}</td>
    <td>{{$client.FileFailures}}</td>
  </tr>{{end}}</tbody></table></div>
  {{else}}<div class="empty-state small"><p>No hay clientes conectados</p></div>{{end}}
  </div>
</div>
{{end}}{{else}}<div class="empty-state"><p>No hay servidores conectados</p></div>{{end}}
</div></div>
</div></div>
<script src="/static/app.js"></script>
<script>setTimeout(function(){location.reload()},5000)</script>
</body></html>`,

"process-snapshots": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Q2PRO Anticheat - Process Snapshots</title>
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
<div class="card"><div class="card-header"><h3>Filtros</h3></div><div class="card-body">
  <form method="GET" action="/process-snapshots" class="filter-form"><div class="form-row">
    <div class="form-group"><label>Nombre Jugador</label><input type="text" name="name" value="{{.PlayerName}}" placeholder="Buscar nombre..."></div>
    <div class="form-group"><label>IP Jugador</label><input type="text" name="player" value="{{.PlayerIP}}" placeholder="Filtrar por IP..."></div>
    <div class="form-group"><label>Desde</label><input type="date" name="from" value="{{.DateFrom}}"></div>
    <div class="form-group"><label>Hasta</label><input type="date" name="to" value="{{.DateTo}}"></div>
    <div class="form-group"><label>&nbsp;</label><button type="submit" class="btn btn-primary">Filtrar</button></div>
  </div></form>
</div></div>
<div class="card"><div class="card-header"><h3>Snapshots ({{.Total}} total)</h3></div><div class="card-body">
{{if .Snapshots}}
<div class="table-responsive"><table class="data-table">
  <thead><tr><th>ID</th><th>Servidor</th><th>Jugador</th><th>IP</th><th>Procesos</th><th>Modulos</th><th>Violaciones</th><th>Fecha</th></tr></thead>
  <tbody>{{range .Snapshots}}
  <tr class="clickable-row" onclick="location.href='/process-snapshots/{{.ID}}'" style="cursor:pointer">
    <td>{{.ID}}</td>
    <td>{{.ServerAddr}}</td>
    <td>{{.PlayerName}}</td>
    <td>{{.PlayerIP}}</td>
    <td>{{.NumProcesses}}</td>
    <td>{{.NumModules}}</td>
    <td>{{if .Violations}}<span class="badge badge-danger violation-count">{{violationsCount .Violations}} violaciones</span><div class="violation-preview" title="{{.Violations}}">{{violationsPreview .Violations}}</div>{{else}}<span class="badge badge-success">Clean</span>{{end}}</td>
    <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
  </tr>
  {{end}}</tbody>
</table></div>
{{if gt .TotalPages 1}}
<div class="pagination">
  {{if gt .Page 1}}<a href="?page={{sub .Page 1}}&name={{.PlayerName}}&player={{.PlayerIP}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Anterior</a>{{end}}
  <span class="page-info">Pagina {{.Page}} de {{.TotalPages}}</span>
  {{if .HasNext}}<a href="?page={{add .Page 1}}&name={{.PlayerName}}&player={{.PlayerIP}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Siguiente</a>{{end}}
</div>
{{end}}
{{else}}<div class="empty-state"><p>No se encontraron process snapshots</p></div>{{end}}
</div></div>
</div></div>
<script src="/static/app.js"></script>
</body></html>`,

"process-snapshot-detail": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Q2PRO Anticheat - Process Snapshot Detail</title>
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
.violation-name { color: #f87171; font-weight: 500; }
.violation-detail { color: #8b95a5; font-size: 12px; }
</style>
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Process Snapshot #{{.Snapshot.ID}}</h2></div>
<div class="content">
<div class="card"><div class="card-header">
  <h3>Detalle del Snapshot</h3>
  <a href="/process-snapshots" class="btn btn-sm">Volver</a>
</div><div class="card-body">
  <div class="snapshot-info">
    <div class="info-item"><div class="info-label">Jugador</div><div class="info-value">{{.Snapshot.PlayerName}}</div></div>
    <div class="info-item"><div class="info-label">IP</div><div class="info-value">{{.Snapshot.PlayerIP}}</div></div>
    <div class="info-item"><div class="info-label">Servidor</div><div class="info-value">{{.Snapshot.ServerAddr}}</div></div>
    <div class="info-item"><div class="info-label">Fecha</div><div class="info-value">{{.Snapshot.Timestamp.Format "2006-01-02 15:04:05"}}</div></div>
    <div class="info-item"><div class="info-label">Total Procesos</div><div class="info-value">{{.Snapshot.NumProcesses}}</div></div>
    <div class="info-item"><div class="info-label">Total Modulos</div><div class="info-value">{{.Snapshot.NumModules}}</div></div>
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

  <div class="section-title">Procesos</div>
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

  <div class="section-title">Modulos</div>
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
        <form method="POST" action="/blacklist" style="display:inline" onsubmit="return confirm('¿Agregar \'{{.Name}}\' a la Blacklist de módulos?')">
          <input type="hidden" name="action" value="add">
          <input type="hidden" name="type" value="module">
          <input type="hidden" name="pattern" value="{{.Name}}">
          <input type="hidden" name="added_by" value="snapshot_inspect">
          <button type="submit" class="btn btn-sm btn-outline-danger" title="Agregar módulo a la blacklist">&#128683; Blacklist</button>
        </form>
      </td>
    </tr>
    {{end}}</tbody>
  </table></div>
  {{else}}<p style="color:#8b95a5">No hay datos de modulos</p>{{end}}
</div></div>
</div></div>
<script src="/static/app.js"></script>
</body></html>`,

"blacklist": `<!DOCTYPE html>
<html lang="es"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Q2PRO Anticheat - Blacklist</title>
<link rel="stylesheet" href="/static/style.css">
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Blacklist de Procesos y Módulos</h2></div>
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
  <div class="stat-card"><div class="stat-icon red">&#9888;</div><div class="stat-info"><div class="stat-value">{{.ProcessCount}}</div><div class="stat-label">Patrones Activos (Procesos)</div></div></div>
  <div class="stat-card"><div class="stat-icon orange">&#128737;</div><div class="stat-info"><div class="stat-value">{{.ModuleCount}}</div><div class="stat-label">Patrones Activos (Módulos)</div></div></div>
  <div class="stat-card"><div class="stat-icon blue">&#128196;</div><div class="stat-info"><div class="stat-value">{{.TotalEntries}}</div><div class="stat-label">Total Patrones Registrados</div></div></div>
</div>

<div class="card"><div class="card-header"><h3>Agregar Patrón a la Blacklist</h3></div><div class="card-body">
  <form method="POST" action="/blacklist" class="filter-form">
    <input type="hidden" name="action" value="add">
    <div class="form-row">
      <div class="form-group" style="max-width: 180px;"><label>Tipo de Elemento</label>
        <select name="type" required>
          <option value="process">Proceso (.exe)</option>
          <option value="module">Módulo (.dll)</option>
        </select>
      </div>
      <div class="form-group" style="flex: 2;"><label>Patrón a Bloquear (coincidencia de texto)</label>
        <input type="text" name="pattern" required placeholder="ej: cheatengine, aimware, xenos.dll, speedhack">
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
    <th>Patrón de Coincidencia</th>
    <th style="width: 110px;">Origen</th>
    <th style="width: 110px;">Estado</th>
    <th style="width: 130px;">Agregado por</th>
    <th style="width: 120px; text-align: right;">Acciones</th>
  </tr></thead>
  <tbody>{{range .Entries}}
  <tr{{if not .Enabled}} class="row-disabled"{{end}} data-type="{{.Type}}" data-status="{{if .Enabled}}active{{else}}inactive{{end}}" data-pattern="{{.Pattern}}">
    <td>{{.ID}}</td>
    <td><span class="badge {{if eq .Type "process"}}badge-danger{{else}}badge-warning{{end}}">{{if eq .Type "process"}}Proceso{{else}}Módulo{{end}}</span></td>
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

"sidebar": `<button class="hamburger" onclick="toggleSidebar()">&#9776;</button>
<div class="sidebar-overlay" onclick="toggleSidebar()"></div>
<div class="sidebar">
  <div class="sidebar-header"><div class="logo">&#128737;</div><span class="logo-text">Anticheat</span></div>
  <nav class="sidebar-nav">
    <a href="/" class="nav-item {{if eq .CurrentPage "dashboard"}}active{{end}}"><span class="nav-icon">&#9632;</span> Dashboard</a>
    <a href="/screenshots" class="nav-item {{if eq .CurrentPage "screenshots"}}active{{end}}"><span class="nav-icon">&#128247;</span> Screenshots</a>
    <a href="/violations" class="nav-item {{if eq .CurrentPage "violations"}}active{{end}}"><span class="nav-icon">&#9888;</span> Violations</a>
    <a href="/process-snapshots" class="nav-item {{if eq .CurrentPage "process-snapshots"}}active{{end}}"><span class="nav-icon">&#128737;</span> Processes</a>
    <a href="/blacklist" class="nav-item {{if eq .CurrentPage "blacklist"}}active{{end}}"><span class="nav-icon">&#128683;</span> Blacklist</a>
    <a href="/servers" class="nav-item {{if eq .CurrentPage "servers"}}active{{end}}"><span class="nav-icon">&#127760;</span> Servers</a>
  </nav>
  <div class="sidebar-footer"><a href="/logout" class="nav-item logout"><span class="nav-icon">&#10140;</span> Salir</a></div>
</div>`,
}
