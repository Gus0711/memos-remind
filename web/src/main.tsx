import "@github/relative-time-element";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import React, { useEffect, useRef } from "react";
import { createRoot } from "react-dom/client";
import { Toaster } from "react-hot-toast";
import { RouterProvider } from "react-router-dom";
import "./i18n";
import "./index.css";
import { ErrorBoundary } from "@/components/ErrorBoundary";
import { AuthProvider, useAuth } from "@/contexts/AuthContext";
import { InstanceProvider, useInstance } from "@/contexts/InstanceContext";
import { ViewProvider } from "@/contexts/ViewContext";
import { queryClient } from "@/lib/query-client";
import router from "./router";
import { applyLocaleEarly } from "./utils/i18n";
import { applyThemeEarly } from "./utils/theme";
import "leaflet/dist/leaflet.css";
import "katex/dist/katex.min.css";

// Apply theme and locale early to prevent flash
applyThemeEarly();
applyLocaleEarly();

// Register service worker for push notifications
const registerServiceWorker = async () => {
  if (!("serviceWorker" in navigator)) {
    console.log("[SW] Service Worker not supported in this browser");
    return;
  }

  console.log("[SW] Registering service worker...");

  try {
    // Use absolute path to ensure correct resolution
    const registration = await navigator.serviceWorker.register("/sw.js", {
      scope: "/",
    });

    console.log("[SW] Service Worker registered successfully:", registration.scope);
    console.log("[SW] Registration state:", registration.active ? "active" : registration.waiting ? "waiting" : "installing");

    // Check for updates periodically
    registration.addEventListener("updatefound", () => {
      const newWorker = registration.installing;
      if (newWorker) {
        console.log("[SW] New Service Worker installing...");
        newWorker.addEventListener("statechange", () => {
          console.log("[SW] Service Worker state changed:", newWorker.state);
          if (newWorker.state === "installed" && navigator.serviceWorker.controller) {
            console.log("[SW] New Service Worker installed, refresh to update");
          }
        });
      }
    });

    // Log when the SW becomes ready
    navigator.serviceWorker.ready.then((reg) => {
      console.log("[SW] Service Worker is ready and active:", reg.scope);
    });
  } catch (error) {
    console.error("[SW] Service Worker registration failed:", error);
    // Log more details about the error
    if (error instanceof Error) {
      console.error("[SW] Error name:", error.name);
      console.error("[SW] Error message:", error.message);
    }
  }
};

// Register SW immediately - don't wait for load event
// This ensures the SW is registered as early as possible
if (document.readyState === "loading") {
  // DOM is still loading, wait for DOMContentLoaded (faster than load)
  document.addEventListener("DOMContentLoaded", registerServiceWorker);
} else {
  // DOM is already ready, register immediately
  registerServiceWorker();
}

// Inner component that initializes contexts
function AppInitializer({ children }: { children: React.ReactNode }) {
  const { isInitialized: authInitialized, initialize: initAuth } = useAuth();
  const { isInitialized: instanceInitialized, initialize: initInstance } = useInstance();
  const initStartedRef = useRef(false);

  // Initialize on mount - run in parallel for better performance
  useEffect(() => {
    if (initStartedRef.current) return;
    initStartedRef.current = true;

    const init = async () => {
      await Promise.all([initInstance(), initAuth()]);
    };
    init();
  }, [initAuth, initInstance]);

  if (!authInitialized || !instanceInitialized) {
    return null;
  }

  return <>{children}</>;
}

function Main() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <InstanceProvider>
          <AuthProvider>
            <ViewProvider>
              <AppInitializer>
                <RouterProvider router={router} />
                <Toaster position="top-right" />
              </AppInitializer>
            </ViewProvider>
          </AuthProvider>
        </InstanceProvider>
        <ReactQueryDevtools initialIsOpen={false} />
      </QueryClientProvider>
    </ErrorBoundary>
  );
}

const container = document.getElementById("root");
const root = createRoot(container as HTMLElement);
root.render(<Main />);
