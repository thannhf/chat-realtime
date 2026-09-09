import { useAuthStore } from "../../stores/auth.store";

interface ProfileModalProps {
    onClose: () => void;
}

function ProfileModal({onClose}: ProfileModalProps) {
    const username = useAuthStore((state) => state.username);
    const avatarUrl = useAuthStore((state) => state.avatarUrl);

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4" onClick={onClose}>
            <div className="w-full max-w-lg overflow-hidden rounded-2xl bg-white shadow-xl" onClick={(event) => event.stopPropagation()}>
                {/* Header */}
                <div className="flex items-center justify-between border-b px-5 py-4">
                    <h2 className="text-xl font-bold">
                        Profile 
                    </h2>
                    <button type="button" onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-full text-gray-500 hover:bg-gray-100 hover:text-black">
                        ✕
                    </button>
                </div>

                {/* Cover + Avatar */}
                <div className="relative">
                    {/* Cover */}
                    <div className="h-36 bg-gradient-to-r from-blue-400 via-purple-400 to-pink-400">
                        <button type="button" className="absolute bottom-3 right-3 flex h-9 w-9 items-center justify-center rounded-full bg-white shadow hover:bg-gray-100">
                            📷
                        </button>
                    </div>

                    {/* Avatar */}
                    <div className="absolute left-1/2 top-20 -translate-x-1/2">
                        {avatarUrl ? (
                            <img src={avatarUrl} alt="Avatar" className="h-32 w-32 rounded-full border-4 border-white object-cover shadow-md" />
                        ) : (
                            <div className="flex h-32 w-32 items-center justify-center rounded-full border-4 border-white bg-gray-300 text-3xl font-semibold shadow-md">
                                ?
                            </div>
                        )}

                        <button className="absolute bottom-1 right-1 flex h-9 w-9 items-center justify-center rounded-full bg-white shadow hover:bg-gray-100">
                            📷
                        </button>
                    </div>
                </div>
                {/* User name */}
                <div className="px-6 pb-6 pt-20 text-center">
                    <h3 className="text-2xl font-bold">
                        {username}
                    </h3>

                    <p className="mt-1 text-sm text-gray-500">
                        ConnectPro User 
                    </p>
                </div>

                {/* Profile Information */}
                <div className="border-t px-6 py-5">
                    <h3 className="mb-4 text-lg font-semibold">
                        Personal Information 
                    </h3>

                    <div className="space-y-4">
                        {/* username */}
                        <div className="flex items-center justify-between rounded-lg border p-3">
                            <div>
                                <p className="text-xs text-gray-500">
                                    Username 
                                </p>

                                <p className="font-medium">
                                    {username || "Not Available"}
                                </p>
                            </div>

                            <button type="button" className="rounded-lg px-3 py-1.5 text-sm font-medium hover:bg-gray-100">
                                Edit 
                            </button>
                        </div>

                        {/* Email */}
                        <div className="flex items-center justify-between rounded-lg border p-3">
                            <div>
                                <p className="text-xs text-gray-500">
                                    Email 
                                </p>

                                <p className="font-medium text-gray-500">
                                    Not Available 
                                </p>
                            </div>

                            <button type="button" className="rounded-lg px-3 py-1.5 text-sm font-medium hover:bg-gray-100">
                                Edit 
                            </button>
                        </div>

                        {/* Bio */}
                        <div className="flex items-center justify-between rounded-lg border p-3">
                            <div>
                                <p className="text-xs text-gray-500">
                                    Bio 
                                </p>
                                <p className="font-medium text-gray-500">
                                    Not Available 
                                </p>
                            </div>
                            <button type="button" className="rounded-lg px-3 py-1.5 text-sm font-medium hover:bg-gray-100">
                                Edit 
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default ProfileModal;