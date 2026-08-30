import apiClient from "../libs/axios";

import type {
    ForgotPasswordRequest,
    ForgotPasswordResponse,
    LoginRequest,
    LoginResponse,
    LogoutResponse,
    RefreshTokenRequest,
    RefreshTokenResponse,
    RegisterRequest,
    RegisterResponse,
    ResetPasswordRequest,
    ResetPasswordResponse,
} from "../types/auth"

export const authApi = {
    login: async (data: LoginRequest): Promise<LoginResponse> => {
        const response = await apiClient.post<LoginResponse>(
            "/api/v1/users/login",
            data, 
        );
        return response.data
    },

    register: async (data: RegisterRequest): Promise<RegisterResponse> => {
        const response = await apiClient.post<RegisterResponse>(
            "api/v1/users/register",
            data
        );
        return response.data;
    },

    refresh: async (data: RefreshTokenRequest): Promise<RefreshTokenResponse> => {
        const response = await apiClient.post<RefreshTokenResponse>(
            "api/v1/users/refresh-token",
            data
        )
        return response.data
    },

    logout: async (): Promise<LogoutResponse> => {
        const response = await apiClient.post<LogoutResponse>(
            "/logout",
        );

        return response.data;
    },

    forgotPassword: async (data: ForgotPasswordRequest): Promise<ForgotPasswordResponse> => {
        const response = await apiClient.post<ForgotPasswordResponse>(
            "/api/v1/users/forgot-password",
            data,
        );
        return response.data;
    },

    resetPassword: async (data: ResetPasswordRequest): Promise<ResetPasswordResponse> => {
        const response = await apiClient.post<ResetPasswordResponse>(
            "/api/v1/users/reset-password", 
            data,
        );

        return response.data;
    },
}