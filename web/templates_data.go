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
    <div class="form-group"><label>Jugador IP</label><input type="text" name="player" value="{{.PlayerIP}}" placeholder="Filtrar por IP..."></div>
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
  {{if gt .Page 1}}<a href="?page={{sub .Page 1}}&player={{.PlayerIP}}&from={{.DateFrom}}&to={{.DateTo}}{{if .Unreviewed}}&unreviewed=1{{end}}" class="btn btn-sm">Anterior</a>{{end}}
  <span class="page-info">Pagina {{.Page}} de {{.TotalPages}}</span>
  {{if lt .Page .TotalPages}}<a href="?page={{add .Page 1}}&player={{.PlayerIP}}&from={{.DateFrom}}&to={{.DateTo}}{{if .Unreviewed}}&unreviewed=1{{end}}" class="btn btn-sm">Siguiente</a>{{end}}
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
    <div class="form-group"><label>Jugador IP</label><input type="text" name="player" value="{{.PlayerIP}}" placeholder="Filtrar por IP..."></div>
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
  {{if gt .Page 1}}<a href="?page={{sub .Page 1}}&player={{.PlayerIP}}&type={{.Type}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Anterior</a>{{end}}
  <span class="page-info">Pagina {{.Page}} de {{.TotalPages}}</span>
  {{if lt .Page .TotalPages}}<a href="?page={{add .Page 1}}&player={{.PlayerIP}}&type={{.Type}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Siguiente</a>{{end}}
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
</head><body>
{{template "sidebar" .}}
<div class="main-content">
<div class="topbar"><h2>Process Snapshots</h2></div>
<div class="content">
<div class="card"><div class="card-header"><h3>Filtros</h3></div><div class="card-body">
  <form method="GET" action="/process-snapshots" class="filter-form"><div class="form-row">
    <div class="form-group"><label>Jugador IP</label><input type="text" name="player" value="{{.PlayerIP}}" placeholder="Filtrar por IP..."></div>
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
  <tr>
    <td>{{.ID}}</td>
    <td>{{.ServerAddr}}</td>
    <td>{{.PlayerName}}</td>
    <td>{{.PlayerIP}}</td>
    <td>{{.NumProcesses}}</td>
    <td>{{.NumModules}}</td>
    <td>{{if .Violations}}<span class="badge badge-danger">{{.Violations}}</span>{{else}}<span class="badge badge-success">Clean</span>{{end}}</td>
    <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
  </tr>
  {{end}}</tbody>
</table></div>
{{if gt .Total 20}}
<div class="pagination">
  {{if gt .CurrentPage 1}}<a href="?page={{sub .CurrentPage 1}}&player={{.PlayerIP}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Anterior</a>{{end}}
  <span class="page-info">Pagina {{.CurrentPage}} de {{.TotalPages}}</span>
  {{if .HasNext}}<a href="?page={{add .CurrentPage 1}}&player={{.PlayerIP}}&from={{.DateFrom}}&to={{.DateTo}}" class="btn btn-sm">Siguiente</a>{{end}}
</div>
{{end}}
{{else}}<div class="empty-state"><p>No se encontraron process snapshots</p></div>{{end}}
</div></div>
</div></div>
<script src="/static/app.js"></script>
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
    <a href="/servers" class="nav-item {{if eq .CurrentPage "servers"}}active{{end}}"><span class="nav-icon">&#127760;</span> Servers</a>
  </nav>
  <div class="sidebar-footer"><a href="/logout" class="nav-item logout"><span class="nav-icon">&#10140;</span> Salir</a></div>
</div>`,
}
