// PWA Service Worker Registration
if ('serviceWorker' in navigator) {
  window.addEventListener('load', function() {
    navigator.serviceWorker.register('/sw.js', { scope: '/' })
      .then(function(reg) {
        console.log('[PWA] Service Worker registrado con éxito:', reg.scope);
      })
      .catch(function(err) {
        console.warn('[PWA] Error al registrar Service Worker:', err);
      });
  });
}

// PWA Install Prompt Support
var deferredInstallPrompt = null;
window.addEventListener('beforeinstallprompt', function(e) {
  e.preventDefault();
  deferredInstallPrompt = e;
  var installBanner = document.getElementById('pwa-install-banner');
  if (installBanner) {
    installBanner.style.display = 'flex';
  }
});

window.addEventListener('appinstalled', function() {
  console.log('[PWA] Aplicación instalada exitosamente');
  var installBanner = document.getElementById('pwa-install-banner');
  if (installBanner) installBanner.style.display = 'none';
  deferredInstallPrompt = null;
});

function triggerPWAInstall() {
  if (deferredInstallPrompt) {
    deferredInstallPrompt.prompt();
    deferredInstallPrompt.userChoice.then(function(choiceResult) {
      if (choiceResult.outcome === 'accepted') {
        console.log('[PWA] El usuario aceptó la instalación');
      }
      deferredInstallPrompt = null;
      var installBanner = document.getElementById('pwa-install-banner');
      if (installBanner) installBanner.style.display = 'none';
    });
  } else {
    // Fallback for iOS or already installed
    var isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent) && !window.MSStream;
    if (isIOS) {
      alert('Para instalar en iPhone/iPad:\n1. Toca el botón Compartir (icono con flecha hacia arriba).\n2. Selecciona "Agregar a pantalla de inicio".');
    } else {
      alert('La aplicación ya está instalada o tu navegador no soporta instalación automática.');
    }
  }
}

function dismissInstallBanner() {
  var installBanner = document.getElementById('pwa-install-banner');
  if (installBanner) installBanner.style.display = 'none';
  sessionStorage.setItem('pwa_prompt_dismissed', '1');
}

// Online / Offline Indicator
window.addEventListener('online', function() {
  showNetworkStatus('Conexión reestablecida', 'success');
});
window.addEventListener('offline', function() {
  showNetworkStatus('Modo sin conexión', 'warning');
});

function showNetworkStatus(msg, type) {
  var el = document.getElementById('network-toast');
  if (!el) {
    el = document.createElement('div');
    el.id = 'network-toast';
    document.body.appendChild(el);
  }
  el.className = 'network-toast toast-' + type + ' show';
  el.textContent = msg;
  setTimeout(function() {
    el.classList.remove('show');
  }, 4000);
}

// Sidebar toggle for mobile
function toggleSidebar() {
  var sidebar = document.querySelector('.sidebar');
  var overlay = document.querySelector('.sidebar-overlay');
  if (sidebar) sidebar.classList.toggle('open');
  if (overlay) overlay.classList.toggle('active');
}

// Lightbox with Zoom, Touch Gestures, Filters & Download
var currentZoom = 1;
var isInverted = false;
var lastTap = 0;
var initialTouchDist = 0;

function openLightbox(src, name, ip, date, server) {
  var lb = document.getElementById('lightbox');
  var img = document.getElementById('lightbox-img');
  var info = document.getElementById('lightbox-info');
  var profileLink = document.getElementById('lightbox-profile-btn');
  var downloadLink = document.getElementById('lightbox-download-btn');

  if (!lb || !img) return;

  img.src = src;
  currentZoom = 1;
  isInverted = false;
  img.style.transform = 'scale(1)';
  img.style.filter = 'none';

  if (info) {
    info.innerHTML = '<strong>' + name + '</strong> (' + ip + ') &mdash; ' + date + ' &mdash; ' + server;
  }
  if (profileLink) {
    profileLink.href = '/player?q=' + encodeURIComponent(name);
  }
  if (downloadLink) {
    downloadLink.href = src;
    downloadLink.download = (name || 'screenshot') + '_' + (date || '').replace(/[: ]/g, '_') + '.webp';
  }

  lb.classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeLightbox() {
  var lb = document.getElementById('lightbox');
  if (lb) lb.classList.remove('active');
  document.body.style.overflow = '';
}

function zoomLightbox(delta) {
  var img = document.getElementById('lightbox-img');
  if (!img) return;
  currentZoom += delta;
  if (currentZoom < 0.5) currentZoom = 0.5;
  if (currentZoom > 4) currentZoom = 4;
  img.style.transform = 'scale(' + currentZoom + ')';
}

function resetLightboxZoom() {
  var img = document.getElementById('lightbox-img');
  if (!img) return;
  currentZoom = 1;
  img.style.transform = 'scale(1)';
}

function toggleLightboxInvert() {
  var img = document.getElementById('lightbox-img');
  if (!img) return;
  isInverted = !isInverted;
  img.style.filter = isInverted ? 'invert(1) hue-rotate(180deg) contrast(1.2)' : 'none';
}

// Mobile Touch Double-Tap to Zoom & Pinch
document.addEventListener('DOMContentLoaded', function() {
  var img = document.getElementById('lightbox-img');
  if (img) {
    img.addEventListener('touchend', function(e) {
      var currentTime = new Date().getTime();
      var tapLength = currentTime - lastTap;
      if (tapLength < 300 && tapLength > 0) {
        // Double tap detected
        e.preventDefault();
        if (currentZoom > 1) {
          resetLightboxZoom();
        } else {
          zoomLightbox(1.0);
        }
      }
      lastTap = currentTime;
    });

    img.addEventListener('touchstart', function(e) {
      if (e.touches.length === 2) {
        initialTouchDist = Math.hypot(
          e.touches[0].pageX - e.touches[1].pageX,
          e.touches[0].pageY - e.touches[1].pageY
        );
      }
    });

    img.addEventListener('touchmove', function(e) {
      if (e.touches.length === 2 && initialTouchDist > 0) {
        var currentDist = Math.hypot(
          e.touches[0].pageX - e.touches[1].pageX,
          e.touches[0].pageY - e.touches[1].pageY
        );
        var diff = (currentDist - initialTouchDist) / 200;
        zoomLightbox(diff);
        initialTouchDist = currentDist;
      }
    });
  }
});

document.addEventListener('keydown', function(e) {
  if (e.key === 'Escape') closeLightbox();
  if (e.key === '+' || e.key === '=') zoomLightbox(0.25);
  if (e.key === '-') zoomLightbox(-0.25);
  if (e.key === '0') resetLightboxZoom();
  if (e.key === 'i' || e.key === 'I') toggleLightboxInvert();
});

// Bulk Screenshots Selection
function toggleSelectAllScreenshots(masterCheckbox) {
  var checkboxes = document.querySelectorAll('.ss-checkbox');
  checkboxes.forEach(function(cb) {
    cb.checked = masterCheckbox.checked;
  });
  updateBulkActionState();
}

function updateBulkActionState() {
  var checked = document.querySelectorAll('.ss-checkbox:checked');
  var countEl = document.getElementById('selected-count');
  var bulkBtn = document.getElementById('bulk-review-btn');
  if (countEl) countEl.textContent = checked.length;
  if (bulkBtn) bulkBtn.disabled = checked.length === 0;
}

function submitBulkReview() {
  var checked = document.querySelectorAll('.ss-checkbox:checked');
  if (checked.length === 0) return;
  var ids = [];
  checked.forEach(function(cb) {
    ids.push(cb.value);
  });
  var input = document.getElementById('bulk-ids-input');
  var form = document.getElementById('bulk-review-form');
  if (input && form) {
    input.value = ids.join(',');
    form.submit();
  }
}

// Format file sizes on page load & check install banner dismissed
document.addEventListener('DOMContentLoaded', function() {
  var el = document.getElementById('total-size');
  if (el) {
    var bytes = parseInt(el.textContent);
    if (!isNaN(bytes)) {
      el.textContent = formatBytes(bytes);
    }
  }

  // Hide install banner if user dismissed previously in this session
  if (sessionStorage.getItem('pwa_prompt_dismissed') === '1') {
    var banner = document.getElementById('pwa-install-banner');
    if (banner) banner.style.display = 'none';
  }
});

function formatBytes(b) {
  if (b >= 1073741824) return (b / 1073741824).toFixed(2) + ' GB';
  if (b >= 1048576) return (b / 1048576).toFixed(1) + ' MB';
  if (b >= 1024) return (b / 1024).toFixed(1) + ' KB';
  return b + ' B';
}
