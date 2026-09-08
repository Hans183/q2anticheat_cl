// Service Worker for Q2PRO Anticheat PWA
const CACHE_NAME = 'q2anticheat-pwa-v2';
const PRECACHE_ASSETS = [
  '/',
  '/static/style.css',
  '/static/app.js',
  '/static/manifest.json',
  '/static/icon.svg',
  '/static/icon-192.png',
  '/static/icon-512.png',
  '/static/badge-monochrome.png'
];

// Install: Cache essential shell assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(PRECACHE_ASSETS).catch((err) => {
        console.warn('[SW] Pre-cache partial warning:', err);
      });
    }).then(() => self.skipWaiting())
  );
});

// Activate: Clean up old cache versions and claim clients
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            console.log('[SW] Deleting old cache:', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    }).then(() => self.clients.claim())
  );
});

// Fetch: Smart caching strategy
self.addEventListener('fetch', (event) => {
  const req = event.request;
  const url = new URL(req.url);

  // Skip non-GET requests and non-http(s) schemes
  if (req.method !== 'GET' || !url.protocol.startsWith('http')) {
    return;
  }

  // 1. Static Assets (CSS, JS, Icons) -> Cache-First with Background Update (Stale-While-Revalidate)
  if (url.pathname.startsWith('/static/')) {
    event.respondWith(
      caches.match(req).then((cachedResp) => {
        const fetchPromise = fetch(req).then((networkResp) => {
          if (networkResp && networkResp.status === 200) {
            const respClone = networkResp.clone();
            caches.open(CACHE_NAME).then((cache) => cache.put(req, respClone));
          }
          return networkResp;
        }).catch(() => cachedResp);

        return cachedResp || fetchPromise;
      })
    );
    return;
  }

  // 2. Navigation / HTML pages -> Network-First with Cache Fallback
  if (req.mode === 'navigate' || req.headers.get('accept')?.includes('text/html')) {
    event.respondWith(
      fetch(req).then((networkResp) => {
        if (networkResp && networkResp.status === 200) {
          const respClone = networkResp.clone();
          caches.open(CACHE_NAME).then((cache) => cache.put(req, respClone));
        }
        return networkResp;
      }).catch(async () => {
        const cachedResp = await caches.match(req);
        if (cachedResp) return cachedResp;
        
        // Fallback to cached root dashboard or offline fallback
        const rootResp = await caches.match('/');
        if (rootResp) return rootResp;

        return new Response(
          '<!DOCTYPE html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Sin Conexión</title><link rel="stylesheet" href="/static/style.css"></head><body style="display:flex;align-items:center;justify-content:center;height:100vh;background:#0f172a;color:#fff;text-align:center;font-family:sans-serif;padding:20px;"><div><h1>📡 Sin Conexión</h1><p>El dashboard requiere conexión para sincronizar los datos en vivo del servidor.</p><button onclick="location.reload()" style="background:#3b82f6;color:#fff;border:none;padding:10px 20px;border-radius:6px;font-size:16px;margin-top:16px;cursor:pointer;">Reintentar</button></div></body></html>',
          { headers: { 'Content-Type': 'text/html; charset=utf-8' } }
        );
      })
    );
    return;
  }

  // 3. All other requests (e.g. Screenshot images) -> Cache-first with network fallback
  event.respondWith(
    caches.match(req).then((cachedResp) => {
      if (cachedResp) return cachedResp;
      return fetch(req).then((networkResp) => {
        if (networkResp && networkResp.status === 200) {
          const respClone = networkResp.clone();
          caches.open(CACHE_NAME).then((cache) => cache.put(req, respClone));
        }
        return networkResp;
      });
    })
  );
});

// 4. Web Push Notification Event Listener
self.addEventListener('push', (event) => {
  let data = {
    title: '⚠️ Violación Anticheat',
    body: 'Se ha detectado una nueva infracción en un servidor.',
    icon: '/static/icon-192.png',
    badge: '/static/badge-monochrome.png',
    url: '/violations'
  };

  if (event.data) {
    try {
      data = Object.assign(data, event.data.json());
    } catch (e) {
      data.body = event.data.text();
    }
  }

  const options = {
    body: data.body,
    icon: data.icon || '/static/icon-192.png',
    badge: data.badge || '/static/badge-monochrome.png',
    vibrate: [200, 100, 200],
    data: {
      url: data.url || '/violations'
    },
    tag: data.tag || 'anticheat-violation',
    renotify: true
  };

  event.waitUntil(
    self.registration.showNotification(data.title, options)
  );
});

// 5. Notification Click Event Listener: Focus or Open Dashboard Page
self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  const targetUrl = (event.notification.data && event.notification.data.url) ? event.notification.data.url : '/violations';

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      for (const client of clientList) {
        if (client.url && 'focus' in client) {
          if (client.navigate) {
            client.navigate(targetUrl);
          }
          return client.focus();
        }
      }
      if (clients.openWindow) {
        return clients.openWindow(targetUrl);
      }
    })
  );
});

