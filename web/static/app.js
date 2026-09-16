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

function arrayBufferToBase64Url(buffer) {
  if (!buffer) return '';
  var binary = '';
  var bytes = new Uint8Array(buffer);
  for (var i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return window.btoa(binary)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

async function forceResubscribePush(keyPublicKey) {
  try {
    const reg = await navigator.serviceWorker.ready;
    let sub = await reg.pushManager.getSubscription();
    if (sub) {
      try {
        await sub.unsubscribe();
      } catch (unsubErr) {
        console.warn('[PUSH] Unsubscribe error (safe to ignore):', unsubErr);
      }
    }

    if (!keyPublicKey) {
      const keyResp = await fetch('/api/push/vapid-key');
      const keyData = await keyResp.json();
      keyPublicKey = keyData ? keyData.publicKey : null;
    }
    if (!keyPublicKey) {
      console.error('[PUSH] No se pudo obtener la clave VAPID pública del servidor.');
      return null;
    }

    // Short delay to allow browser push service unregistration to settle
    await new Promise(resolve => setTimeout(resolve, 150));

    const applicationServerKey = urlBase64ToUint8Array(keyPublicKey);
    sub = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: applicationServerKey
    });

    const subJson = sub.toJSON();
    const saveResp = await fetch('/api/push/subscribe', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        endpoint: subJson.endpoint,
        keys: subJson.keys
      })
    });

    if (saveResp.ok) {
      localStorage.setItem('ac_vapid_key', keyPublicKey);
      localStorage.setItem('ac_push_subscribed', 'true');
      updatePushUI('subscribed');
      console.log('[PUSH] Suscripción renovada con éxito con la clave VAPID actual.');
      return sub;
    } else {
      console.error('[PUSH] Error al registrar suscripción en la base de datos.');
      return null;
    }
  } catch (err) {
    console.error('[PUSH] Error en forceResubscribePush:', err);
    return null;
  }
}

async function checkPushSubscriptionState() {
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
    updatePushUI('unsupported');
    return;
  }

  try {
    const reg = await navigator.serviceWorker.ready;
    const sub = await reg.pushManager.getSubscription();

    // Fetch server VAPID key
    const keyResp = await fetch('/api/push/vapid-key');
    const keyData = await keyResp.json();
    const serverKey = keyData ? keyData.publicKey : null;

    if (!serverKey) {
      if (sub) updatePushUI('subscribed');
      return;
    }

    const storedKey = localStorage.getItem('ac_vapid_key');
    const isPermissionGranted = (Notification.permission === 'granted');

    if (sub) {
      // If server VAPID key rotated since last subscription, auto-heal
      if (storedKey && storedKey !== serverKey && isPermissionGranted) {
        console.warn('[PUSH] Clave VAPID del servidor cambió (' + storedKey + ' -> ' + serverKey + '). Auto-reparando suscripción...');
        await forceResubscribePush(serverKey);
        return;
      }

      updatePushUI('subscribed');
      localStorage.setItem('ac_vapid_key', serverKey);
      localStorage.setItem('ac_push_subscribed', 'true');

      // Sync endpoint to DB to ensure it exists in case of DB restart
      const subJson = sub.toJSON();
      if (subJson && subJson.endpoint && subJson.keys) {
        fetch('/api/push/subscribe', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            endpoint: subJson.endpoint,
            keys: subJson.keys
          })
        }).catch(function(err) {
          console.warn('[PUSH] Re-sync warning:', err);
        });
      }
    } else {
      // No active subscription in browser
      if (isPermissionGranted && localStorage.getItem('ac_push_subscribed') === 'true') {
        console.log('[PUSH] Restaurando suscripción push...');
        await forceResubscribePush(serverKey);
      } else {
        updatePushUI('unsubscribed');
      }
    }
  } catch (err) {
    console.warn('[PUSH] Error comprobando estado de suscripción:', err);
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
      localStorage.removeItem('ac_push_subscribed');
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

      const newSub = await forceResubscribePush();
      if (newSub) {
        showNetworkStatus('🔔 Notificaciones Push activadas correctamente', 'success');
      } else {
        throw new Error('No se pudo completar la suscripción.');
      }
    }
  } catch (err) {
    console.error('[PUSH] Error al cambiar suscripción:', err);
    alert('Error al gestionar las notificaciones Push: ' + err.message);
  }
}

async function sendTestPushNotification() {
  try {
    showNetworkStatus('Enviando notificación de prueba...', 'info');
    let resp = await fetch('/api/push/test', { method: 'POST' });
    let data = await resp.json();

    // If 403 Forbidden, 401 or invalid credentials occurred, auto-heal and retry once
    if (!resp.ok && data.error && (data.error.includes('403') || data.error.includes('401') || data.error.includes('credentials') || data.error.includes('No hay dispositivos'))) {
      console.warn('[PUSH] Error en prueba (' + data.error + '). Reparando suscripción automáticamente...');
      showNetworkStatus('Actualizando clave de suscripción...', 'warning');
      const newSub = await forceResubscribePush();
      if (newSub) {
        await new Promise(r => setTimeout(r, 300));
        resp = await fetch('/api/push/test', { method: 'POST' });
        data = await resp.json();
      }
    }

    if (resp.ok && data.ok) {
      showNetworkStatus('🔔 ¡Notificación de prueba enviada exitosamente!', 'success');
    } else {
      const errMsg = data.error || 'No se pudo entregar la notificación de prueba.';
      // Try one more clean auto-resubscribe in background
      await forceResubscribePush();
      alert('⚠️ Error enviando notificación de prueba:\n' + errMsg + '\n\nSe ha renovado la clave en este dispositivo. Intenta enviar la prueba nuevamente.');
      showNetworkStatus('Suscripción renovada. Intenta de nuevo.', 'warning');
    }
  } catch (err) {
    alert('Error enviando prueba: ' + err.message);
  }
}

async function resetAllPushSubscriptions() {
  if (!confirm('¿Desea limpiar todas las suscripciones push de la base de datos? Los dispositivos deberán re-suscribirse.')) {
    return;
  }
  try {
    const resp = await fetch('/api/push/reset', { method: 'POST' });
    const data = await resp.json();
    if (data.ok) {
      // Re-subscribe current device
      await forceResubscribePush();
      showNetworkStatus('Suscripciones restablecidas y dispositivo re-vinculado', 'success');
      setTimeout(() => location.reload(), 1000);
    } else {
      alert('Error: ' + (data.error || 'No se pudo restablecer'));
    }
  } catch (err) {
    alert('Error al restablecer: ' + err.message);
  }
}

function copyVapidEnvVars() {
  const codeElem = document.getElementById('vapid-env-snippet');
  if (!codeElem) return;
  const text = codeElem.innerText || codeElem.textContent;
  navigator.clipboard.writeText(text).then(function() {
    showNetworkStatus('📋 Variables de entorno VAPID copiadas al portapapeles', 'success');
  }).catch(function() {
    alert('No se pudo copiar automáticamente. Por favor copia el texto manualmente.');
  });
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

