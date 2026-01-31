import { useCallback, useEffect, useState } from "react";
import { instanceServiceClient, userServiceClient } from "@/connect";
import useCurrentUser from "./useCurrentUser";

export interface WebPushState {
  supported: boolean;
  permission: NotificationPermission;
  subscribed: boolean;
  enabled: boolean;
  loading: boolean;
  error: string | null;
}

const useWebPush = () => {
  const currentUser = useCurrentUser();
  const [state, setState] = useState<WebPushState>({
    supported: false,
    permission: "default",
    subscribed: false,
    enabled: false,
    loading: true,
    error: null,
  });

  // Check if web push is supported and get initial state
  useEffect(() => {
    const checkSupport = async () => {
      console.log("[WebPush] checkSupport: starting...");

      // Check browser support
      const supported = "serviceWorker" in navigator && "PushManager" in window && "Notification" in window;
      console.log("[WebPush] checkSupport: browser supported =", supported);

      if (!supported) {
        setState((prev) => ({
          ...prev,
          supported: false,
          loading: false,
        }));
        return;
      }

      // Get notification permission
      const permission = Notification.permission;
      console.log("[WebPush] checkSupport: current permission =", permission);

      // Check if server has web push enabled
      let enabled = false;
      try {
        const response = await instanceServiceClient.getVAPIDPublicKey({});
        enabled = !!response.publicKey;
        console.log("[WebPush] checkSupport: server VAPID enabled =", enabled);
      } catch (error) {
        console.log("[WebPush] checkSupport: VAPID key fetch failed:", error);
      }

      // Check if user is subscribed (with timeout to avoid blocking)
      let subscribed = false;
      if (enabled && currentUser && permission === "granted") {
        try {
          // Use a timeout to avoid blocking forever if SW is not ready
          const registration = await Promise.race([
            navigator.serviceWorker.ready,
            new Promise<null>((_, reject) => setTimeout(() => reject(new Error("SW timeout")), 3000)),
          ]);

          if (registration) {
            console.log("[WebPush] checkSupport: Service Worker ready");
            const subscription = await (registration as ServiceWorkerRegistration).pushManager.getSubscription();
            subscribed = !!subscription;
            console.log("[WebPush] checkSupport: existing subscription =", subscribed);
          }
        } catch (error) {
          console.log("[WebPush] checkSupport: SW not ready yet, skipping subscription check");
        }
      }

      console.log("[WebPush] checkSupport: complete. Setting state with loading=false");
      setState({
        supported,
        permission,
        subscribed,
        enabled,
        loading: false,
        error: null,
      });
    };

    checkSupport();
  }, [currentUser]);

  // Subscribe to push notifications
  const subscribe = useCallback(async () => {
    console.log("[WebPush] subscribe() called, currentUser =", currentUser?.name);

    if (!currentUser) {
      console.log("[WebPush] subscribe: no currentUser, aborting");
      setState((prev) => ({ ...prev, error: "Not authenticated", loading: false }));
      return false;
    }

    console.log("[WebPush] subscribe: setting loading=true");
    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      // Step 1: Request notification permission
      console.log("[WebPush] Step 1: Calling Notification.requestPermission()...");
      const permission = await Notification.requestPermission();
      console.log("[WebPush] Permission result:", permission);

      if (permission !== "granted") {
        setState((prev) => ({
          ...prev,
          permission,
          loading: false,
          error: "Notification permission denied",
        }));
        return false;
      }

      // Update permission state immediately
      setState((prev) => ({ ...prev, permission: "granted" }));

      // Step 2: Get VAPID public key from server
      console.log("[WebPush] Step 2: Getting VAPID public key...");
      const vapidResponse = await instanceServiceClient.getVAPIDPublicKey({});
      if (!vapidResponse.publicKey) {
        setState((prev) => ({
          ...prev,
          loading: false,
          error: "Web push not enabled on server",
        }));
        return false;
      }
      console.log("[WebPush] VAPID key received");

      // Step 3: Wait for service worker with timeout
      console.log("[WebPush] Step 3: Waiting for Service Worker...");
      const registration = await waitForServiceWorker(10000); // 10 second timeout
      console.log("[WebPush] Service Worker ready:", registration.scope);

      // Step 4: Subscribe to push notifications
      console.log("[WebPush] Step 4: Subscribing to push manager...");
      const vapidKey = urlBase64ToUint8Array(vapidResponse.publicKey);
      const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: vapidKey.buffer as ArrayBuffer,
      });
      console.log("[WebPush] Push subscription created:", subscription.endpoint);

      // Step 5: Extract keys and save to server
      console.log("[WebPush] Step 5: Saving subscription to server...");
      const p256dh = arrayBufferToBase64(subscription.getKey("p256dh"));
      const auth = arrayBufferToBase64(subscription.getKey("auth"));

      await userServiceClient.createUserPushSubscription({
        parent: currentUser.name,
        subscription: {
          endpoint: subscription.endpoint,
          p256dh,
          auth,
          userAgent: navigator.userAgent,
        },
      });
      console.log("[WebPush] Subscription saved to server successfully");

      setState((prev) => ({
        ...prev,
        permission: "granted",
        subscribed: true,
        loading: false,
        error: null,
      }));

      return true;
    } catch (error) {
      console.error("[WebPush] Failed to subscribe:", error);
      setState((prev) => ({
        ...prev,
        loading: false,
        error: error instanceof Error ? error.message : "Failed to subscribe",
      }));
      return false;
    }
  }, [currentUser]);

  // Unsubscribe from push notifications
  const unsubscribe = useCallback(async () => {
    if (!currentUser) {
      return false;
    }

    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      // Get current subscription
      const registration = await navigator.serviceWorker.ready;
      const subscription = await registration.pushManager.getSubscription();

      if (subscription) {
        // Unsubscribe from browser
        await subscription.unsubscribe();

        // Find and delete subscription from server
        const response = await userServiceClient.listUserPushSubscriptions({
          parent: currentUser.name,
        });

        for (const sub of response.subscriptions) {
          if (sub.endpoint === subscription.endpoint) {
            await userServiceClient.deleteUserPushSubscription({
              name: sub.name,
            });
            break;
          }
        }
      }

      setState((prev) => ({
        ...prev,
        subscribed: false,
        loading: false,
        error: null,
      }));

      return true;
    } catch (error) {
      console.error("Failed to unsubscribe from push notifications:", error);
      setState((prev) => ({
        ...prev,
        loading: false,
        error: error instanceof Error ? error.message : "Failed to unsubscribe",
      }));
      return false;
    }
  }, [currentUser]);

  return {
    ...state,
    subscribe,
    unsubscribe,
  };
};

// Wait for service worker with timeout
async function waitForServiceWorker(timeoutMs: number): Promise<ServiceWorkerRegistration> {
  // First, check if there's already an active registration
  const registrations = await navigator.serviceWorker.getRegistrations();
  for (const reg of registrations) {
    if (reg.active) {
      return reg;
    }
  }

  // If no active registration, try to register and wait
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      reject(new Error("Service Worker registration timeout"));
    }, timeoutMs);

    navigator.serviceWorker.ready
      .then((registration) => {
        clearTimeout(timeout);
        resolve(registration);
      })
      .catch((error) => {
        clearTimeout(timeout);
        reject(error);
      });

    // Also try to register if not already done
    if (!navigator.serviceWorker.controller) {
      navigator.serviceWorker
        .register("/sw.js", { scope: "/" })
        .then(() => {
          console.log("[WebPush] Service Worker registered during subscribe");
        })
        .catch((err) => {
          console.error("[WebPush] Failed to register SW:", err);
        });
    }
  });
}

// Convert a base64 string to Uint8Array (for VAPID key)
function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}

// Convert ArrayBuffer to base64 string
function arrayBufferToBase64(buffer: ArrayBuffer | null): string {
  if (!buffer) return "";
  const bytes = new Uint8Array(buffer);
  let binary = "";
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return window.btoa(binary);
}

export default useWebPush;
