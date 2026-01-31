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
      // Check browser support
      const supported = "serviceWorker" in navigator && "PushManager" in window && "Notification" in window;

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

      // Check if server has web push enabled
      let enabled = false;
      try {
        const response = await instanceServiceClient.getVAPIDPublicKey({});
        enabled = !!response.publicKey;
      } catch {
        // Web push not enabled on server
      }

      // Check if user is subscribed
      let subscribed = false;
      if (enabled && currentUser) {
        try {
          // Wait for service worker to be ready (registered in main.tsx)
          const registration = await navigator.serviceWorker.ready;
          console.log("[WebPush] Service Worker ready:", registration.scope);

          // Check existing subscription
          const subscription = await registration.pushManager.getSubscription();
          subscribed = !!subscription;
          console.log("[WebPush] Existing subscription:", subscribed ? subscription?.endpoint : "none");
        } catch (error) {
          console.error("[WebPush] Error checking subscription:", error);
        }
      }

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
    if (!currentUser) {
      setState((prev) => ({ ...prev, error: "Not authenticated" }));
      return false;
    }

    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      // Request notification permission
      const permission = await Notification.requestPermission();
      if (permission !== "granted") {
        setState((prev) => ({
          ...prev,
          permission,
          loading: false,
          error: "Notification permission denied",
        }));
        return false;
      }

      // Get VAPID public key from server
      const vapidResponse = await instanceServiceClient.getVAPIDPublicKey({});
      if (!vapidResponse.publicKey) {
        setState((prev) => ({
          ...prev,
          loading: false,
          error: "Web push not enabled on server",
        }));
        return false;
      }

      // Wait for service worker to be ready (registered in main.tsx)
      console.log("[WebPush] Waiting for Service Worker...");
      const registration = await navigator.serviceWorker.ready;
      console.log("[WebPush] Service Worker ready for subscription:", registration.scope);

      // Subscribe to push notifications
      console.log("[WebPush] Subscribing to push manager...");
      const vapidKey = urlBase64ToUint8Array(vapidResponse.publicKey);
      const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: vapidKey.buffer as ArrayBuffer,
      });
      console.log("[WebPush] Push subscription created:", subscription.endpoint);

      // Extract keys from subscription
      const p256dh = arrayBufferToBase64(subscription.getKey("p256dh"));
      const auth = arrayBufferToBase64(subscription.getKey("auth"));

      // Save subscription to server
      console.log("[WebPush] Saving subscription to server...");
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
      console.error("Failed to subscribe to push notifications:", error);
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
