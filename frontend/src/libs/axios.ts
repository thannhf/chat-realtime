import axios, { type InternalAxiosRequestConfig } from "axios";

type RetryableRequestConfig = InternalAxiosRequestConfig & {
  _retry?: boolean;
};

const apiClient = axios.create({
  baseURL: "http://localhost:8000",
  headers: {
    "Content-Type": "application/json",
  },
});

const refreshClient = axios.create({
  baseURL: "http://localhost:8000",
  headers: {
    "Content-Type": "application/json",
  },
});

const refreshAccessToken = async (): Promise<string> => {
  const authStorage = localStorage.getItem("auth-storage");

  if (!authStorage) {
    throw new Error("Auth storage not found");
  }

  const authState = JSON.parse(authStorage);
  const refreshToken = authState?.state?.refreshToken;

  if (!refreshToken) {
    throw new Error("Refresh token not found");
  }

  const response = await refreshClient.post<{
    access_token: string;
    refresh_token: string;
  }>("/api/v1/users/refresh-token", {
    refresh_token: refreshToken,
  });

  const newAccessToken = response.data.access_token;
  const newRefreshToken = response.data.refresh_token;
  const currentState = JSON.parse(
    localStorage.getItem("auth-storage") ?? "{}",
  );

  currentState.state = {
    ...currentState.state,
    accessToken: newAccessToken,
    refreshToken: newRefreshToken,
    isAuthenticated: true,
  };

  localStorage.setItem("auth-storage", JSON.stringify(currentState));

  const { useAuthStore } = await import("../stores/auth.store");

  useAuthStore.setState({
    accessToken: newAccessToken,
    refreshToken: newRefreshToken,
    isAuthenticated: true,
  });

  return newAccessToken;
};

apiClient.interceptors.request.use((config) => {
  const authStorage = localStorage.getItem("auth-storage");
  if (authStorage) {
    try {
      const authState = JSON.parse(authStorage);
      const accessToken = authState?.state?.accessToken;

      if (accessToken) {
        config.headers.Authorization = `Bearer ${accessToken}`;
      }
    } catch {
      // Ignore invalid auth storage
    }
  }

  return config;
},(error) => Promise.reject(error));

let refreshPromise: Promise<string> | null = null;

apiClient.interceptors.response.use((response) => {
  return response;
}, async (error) => {
    const originalRequest = error.config as RetryableRequestConfig;

    if (error.response?.status !== 401 || originalRequest?._retry) {
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    try {
      if (!refreshPromise) {
        refreshPromise = refreshAccessToken();
      }

      const newAccessToken = await refreshPromise;
      originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;

      return apiClient(originalRequest);
    } catch (refreshError) {
      const { useAuthStore } = await import("../stores/auth.store");
      useAuthStore.getState().clearAuth();
      return Promise.reject(refreshError);
    } finally {
      refreshPromise = null;
    }
  },
);

export { refreshClient };
export default apiClient;