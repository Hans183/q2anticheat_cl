// PWA Service Worker Registration & Auto-Update Lifecycle
if ('serviceWorker' in navigator) {
  var swRegistration = null;
  var isRefreshing = false;

  // Auto-reload when new Service Worker takes control
  navigator.serviceWorker.addEventListener('controllerchange', function() {
    if (!isRefreshing) {
      isRefreshing = true;
      console.log('[PWA] Nuevo Service Worker activo, recargando aplicación...');
      window.location.reload();
    }
  });

  function registerServiceWorker() {
    navigator.serviceWorker.register('/sw.js', { scope: '/' })
      .then(function(reg) {
        swRegistration = reg;
        console.log('[PWA] Service Worker registrado con éxito:', reg.scope);

        // Force immediate check for SW update on server
        reg.update().catch(function() {});

        // Check for updates
        reg.addEventListener('updatefound', function() {
          var newWorker = reg.installing;
          if (newWorker) {
            newWorker.addEventListener('statechange', function() {
              if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                console.log('[PWA] Nueva versión detectada. Enviando SKIP_WAITING...');
                newWorker.postMessage({ type: 'SKIP_WAITING' });
              }
            });
          }
        });

        // If worker is already waiting, tell it to take control
        if (reg.waiting && navigator.serviceWorker.controller) {
          reg.waiting.postMessage({ type: 'SKIP_WAITING' });
        }
      })
      .catch(function(err) {
        console.warn('[PWA] Error al registrar Service Worker:', err);
      });
  }

  window.addEventListener('load', registerServiceWorker);

// Global Helper to Force Reset PWA Cache & Service Worker
async function forceResetPWA() {
  if (confirm('¿Deseas desinstalar el Service Worker en caché y forzar la actualización de la PWA?')) {
    if ('serviceWorker' in navigator) {
      const regs = await navigator.serviceWorker.getRegistrations();
      for (let r of regs) {
        await r.unregister();
      }
    }
    if ('caches' in window) {
      const keys = await caches.keys();
      for (let k of keys) {
        await caches.delete(k);
      }
    }
    alert('Caché eliminada con éxito. La página se recargará.');
    window.location.reload(true);
  }
}

  // Mobile Lifecycle: Check for Service Worker updates when returning to foreground
  document.addEventListener('visibilitychange', function() {
    if (document.visibilityState === 'visible' && swRegistration) {
      swRegistration.update().catch(function(err) {
        console.log('[PWA] Check update:', err);
      });
    }
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

// Web Push Notification Support
function urlBase64ToUint8Array(base64String) {
  var padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  var base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  var rawData = window.atob(base64);
  var outputArray = new Uint8Array(rawData.length);
  for (var i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}

async function checkPushSubscriptionState() {
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
    updatePushUI('unsupported');
    return;
  }

  try {
    const reg = await navigator.serviceWorker.ready;
    const sub = await reg.pushManager.getSubscription();
    if (sub) {
      updatePushUI('subscribed');
    } else {
      updatePushUI('unsubscribed');
    }
  } catch (err) {
    console.warn('[PUSH] Error checking subscription:', err);
    updatePushUI('unsubscribed');
  }
}

async function togglePushSubscription() {
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
    alert('Tu navegador no soporta Notificaciones Push.');
    return;
  }

  try {
    const reg = await navigator.serviceWorker.ready;
    let sub = await reg.pushManager.getSubscription();

    if (sub) {
      await sub.unsubscribe();
      await fetch('/api/push/unsubscribe', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ endpoint: sub.endpoint })
      });
      updatePushUI('unsubscribed');
      showNetworkStatus('Notificaciones Push desactivadas', 'warning');
    } else {
      const permission = await Notification.requestPermission();
      if (permission !== 'granted') {
        alert('Se requieren permisos de notificación para habilitar las alertas Push.');
        return;
      }

      const keyResp = await fetch('/api/push/vapid-key');
      const keyData = await keyResp.json();
      if (!keyData.publicKey) {
        throw new Error('No se pudo obtener la clave pública VAPID.');
      }

      const applicationServerKey = urlBase64ToUint8Array(keyData.publicKey);
      sub = await reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: applicationServerKey
      });

      const subJson = sub.toJSON();
      await fetch('/api/push/subscribe', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          endpoint: subJson.endpoint,
          keys: subJson.keys
        })
      });

      updatePushUI('subscribed');
      showNetworkStatus('🔔 Notificaciones Push activadas correctamente', 'success');
    }
  } catch (err) {
    console.error('[PUSH] Error al cambiar suscripción:', err);
    alert('Error al gestionar las notificaciones Push: ' + err.message);
  }
}

async function sendTestPushNotification() {
  try {
    const resp = await fetch('/api/push/test', { method: 'POST' });
    const data = await resp.json();
    if (data.ok) {
      showNetworkStatus('Notificación de prueba enviada', 'success');
    } else {
      alert('Error enviando notificación de prueba');
    }
  } catch (err) {
    alert('Error: ' + err.message);
  }
}

function updatePushUI(state) {
  var bellBtns = document.querySelectorAll('.push-toggle-btn');
  bellBtns.forEach(function(btn) {
    if (state === 'subscribed') {
      btn.classList.add('active');
      btn.title = 'Notificaciones Push activadas (Clic para desactivar)';
      btn.setAttribute('aria-label', 'Notificaciones activadas');
    } else if (state === 'unsubscribed') {
      btn.classList.remove('active');
      btn.title = 'Activar Notificaciones Push de violaciones';
      btn.setAttribute('aria-label', 'Activar notificaciones');
    } else if (state === 'unsupported') {
      btn.disabled = true;
      btn.title = 'Notificaciones Push no soportadas en este navegador';
    }
  });

  var pushStatusText = document.getElementById('push-status-text');
  if (pushStatusText) {
    if (state === 'subscribed') {
      pushStatusText.innerHTML = '<span style="color:#10b981;font-weight:bold;">🟢 Activadas</span>';
    } else if (state === 'unsubscribed') {
      pushStatusText.innerHTML = '<span style="color:#ef4444;font-weight:bold;">🔴 Desactivadas</span>';
    } else {
      pushStatusText.innerHTML = '<span style="color:#6b7280;">⚠️ No soportadas</span>';
    }
  }
}

document.addEventListener('DOMContentLoaded', function() {
  checkPushSubscriptionState();
});

