import apiClient from "../libs/axios";
import type { PublicUser } from "../types/user";

interface GetUsersByIDsResponse {
    data: PublicUser[];
}

export const userApi = {
    getUsersByIDs: async (userIds: string[]): Promise<PublicUser[]> => {
        if (userIds.length === 0) {
            return []
        }

        const response = await apiClient.get<GetUsersByIDsResponse>(
            "/api/v1/users",
            {params: {ids: userIds.join(",")}}
        )
        // console.log("FETCHED USERS:", response);
        return response.data.data;
    },

    uploadAvatar: async (file: File): Promise<string> => {
        const formData = new FormData();
        formData.append("avatar", file);

        const response = await apiClient.post<{avatar_url: string}>(
            "/api/v1/users/avatar/upload", formData,
        );

        return response.data.avatar_url;
    }
}