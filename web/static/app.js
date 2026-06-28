function openLightbox(src, name, ip, date, server) {
  var lb = document.getElementById('lightbox');
  var img = document.getElementById('lightbox-img');
  var info = document.getElementById('lightbox-info');
  img.src = src;
  info.innerHTML = '<strong>' + name + '</strong> (' + ip + ') &mdash; ' + date + ' &mdash; ' + server;
  lb.classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeLightbox() {
  var lb = document.getElementById('lightbox');
  lb.classList.remove('active');
  document.body.style.overflow = '';
}

document.addEventListener('keydown', function(e) {
  if (e.key === 'Escape') closeLightbox();
});

// Format file sizes on page load
document.addEventListener('DOMContentLoaded', function() {
  var el = document.getElementById('total-size');
  if (el) {
    var bytes = parseInt(el.textContent);
    if (!isNaN(bytes)) {
      el.textContent = formatBytes(bytes);
    }
  }
});

function formatBytes(b) {
  if (b >= 1048576) return (b / 1048576).toFixed(1) + ' MB';
  if (b >= 1024) return (b / 1024).toFixed(1) + ' KB';
  return b + ' B';
}
