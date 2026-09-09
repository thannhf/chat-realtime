import { useNavigate } from "react-router";
import { useState } from "react";
import { useAuthStore } from "../../stores/auth.store";
import ProfileModal from "../profile/ProfileModal";

function AppHeader() {
    const logout = useAuthStore((state) => state.logout);
    const avatarUrl = useAuthStore((state) => state.avatarUrl);
    const username = useAuthStore((state) => state.username);
    const navigate = useNavigate();

    const [isMenuOpen, setIsMenuOpen] = useState(false);
    const [isProfileOpen, setIsProfileOpen] = useState(false);

    const handleLogout = async () => {
        await logout();
        navigate("/login", { replace: true });
    };

    const handleProfile = () => {
        setIsMenuOpen(false);
        setIsProfileOpen(true);
    }

    return (
        <header className="flex h-16 items-center border-b px-6">
            <h1 className="text-xl font-bold">
                ConnectPro
            </h1>

            <div className="relative ml-auto">
                {/* account */}
                <button type="button" onClick={() => setIsMenuOpen((prev) => !prev)} className="flex items-center gap-2 rounded-full p-1 hover:bg-gray-100">
                    {avatarUrl ? (
                        <img src={avatarUrl} alt="Avatar" className="h-10 w-10 rounded-full object-cover" />
                    ) : (
                        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gray-300 text-sm font-semibold">
                            ?
                        </div>
                    )}
                </button>

                {/* account menu */}
                {isMenuOpen && (
                    <div className="absolute right-0 top-12 z-50 w-56 rounded-xl border bg-white p-2 shadow-lg">
                        <div className="border-b px-3 py-3">
                            <p className="font-semibold">
                                {username}
                            </p>
                        </div>

                        <button type="button" onClick={handleProfile} className="w-full rounded-lg px-3 py-2 text-left hover:bg-gray-100">
                            Profile
                        </button>

                        <button type="button" onClick={handleLogout} className="w-full rounded-lg px-3 py-2 text-left text-red-600 hover:bg-red-50">
                            Logout 
                        </button>
                    </div>
                )}

                {isProfileOpen && (
                    <ProfileModal onClose={() => setIsProfileOpen(false)} />
                )}
            </div>
        </header>
    );
}

export default AppHeader;