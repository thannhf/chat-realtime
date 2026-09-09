import { Navigate, Outlet } from "react-router";
import { useAuthStore } from "../stores/auth.store";


export default function PublicRoute() {
    const isAuthenticated = useAuthStore(
        (state) => state.isAuthenticated,
    );

    if(isAuthenticated) {
        return <Navigate to="/chat" replace />
    }

    return <Outlet />
}