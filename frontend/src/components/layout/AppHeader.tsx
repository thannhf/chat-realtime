import { useNavigate } from "react-router";
import { useAuthStore } from "../../stores/auth.store"


function AppHeader() {
    const logout = useAuthStore((state) => state.logout);
    const navigate = useNavigate();

    const handleLogout = async () => {
        await logout();
        navigate("/login", {replace: true});
    }

    return (
        <header className="flex h-16 items-center border-b px-6">
            <h1 className="text-xl font-bold">
                ConnectPro
            </h1>
            <button onClick={handleLogout}>
                Logout 
            </button>
        </header>
    )
}

export default AppHeader