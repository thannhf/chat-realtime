import { create } from "zustand"
import { persist } from "zustand/middleware";
import { authApi } from "../api/auth.api"

interface AuthState {
    username: string | null;
    userId: string | null;
    accessToken: string | null;
    refreshToken: string | null;
    isAuthenticated: boolean;
    hydrated: boolean;

    login: (username: string, password: string) => Promise<void>;
    logout: () => Promise<void>;
    clearAuth: () => void;
    refresh: () => Promise<string>;
    register: (username: string, password: string) => Promise<void>;
    forgotPassword: (username: string) => Promise<void>;
    resetPassword: (username: string, otp: string, newPassword: string) => Promise<void>;
    setHydrated: () => void;
}

export const useAuthStore = create<AuthState>()(
    persist((set, get) => ({
        username: null,
        userId: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        hydrated: false,

        login: async (username, password) => {
            const response = await authApi.login({
                username,
                password,
            });
            set({
                username,
                userId: response.user.id,
                accessToken: response.access_token,
                refreshToken: response.refresh_token,
                isAuthenticated: true,
            });
        },

        refresh: async () => {
            const currentRefreshToken = get().refreshToken;
            
            if(!currentRefreshToken) {
                throw new Error("Refresh token not found");
            }

            const response = await authApi.refresh({
                refresh_token: currentRefreshToken,
            });

            set({
                accessToken: response.access_token,
                refreshToken: response.refresh_token,
                isAuthenticated: true,
            });

            return response.access_token;
        },

        logout: async () => {
            try {
                await authApi.logout();
            } finally {
                get().clearAuth();
            }
        },

        clearAuth: () => {
            set({
                username: null,
                userId: null,
                accessToken: null,
                refreshToken: null,
                isAuthenticated: false,
            });
        },

        register: async (username, password) => {
            await authApi.register({
                username,
                password,
            });
        },

        forgotPassword: async (username) => {
            await authApi.forgotPassword({username});
        },

        resetPassword: async (username, otp, newPassword) => {
            await authApi.resetPassword({
                username, otp, newPassword,
            });
        },

        setHydrated:() => {
            set({
                hydrated: true,
            })
        },
    }),
    {
        name: "auth-storage",

        onRehydrateStorage: () => (state) => {
            state?.setHydrated();
        }
    },
));