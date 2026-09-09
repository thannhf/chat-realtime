import { userApi } from "../api/user.api";
import type { PublicUser } from "../types/user";
import { create } from "zustand";

interface UserState {
    users: Record<string, PublicUser>
    fetchUsers: (userIds: string[]) => Promise<void>;
    
}

export const userUserStore = create<UserState>((set) => ({
    users: {},

    fetchUsers: async (userIds) => {
        if (userIds.length === 0) {
            return;
        }

        const users = await userApi.getUsersByIDs(userIds);
        // console.log("USER API RESULT:", users);
        set((state) => {
            const updatedUsers = {...state.users};

            for (const user of users) {
                // console.log("ADDING USER TO STORE:", user);
                updatedUsers[user.id] = user;
            }
            // console.log("UPDATED USERS:", updatedUsers);
            return {
                users: updatedUsers,
            }
        })
    }
}))