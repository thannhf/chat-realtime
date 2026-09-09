import { Link } from "react-router";
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { loginSchema, type LoginFormData } from "../../types/auth";
import { useAuthStore } from "../../stores/auth.store";
import { useNavigate } from "react-router";
import { useState } from "react";

export default function LoginForm() {
    const [loading, setLoading] = useState(false)
    const navigate = useNavigate();
    const {register, handleSubmit, formState: {errors}} =
    useForm<LoginFormData>({
        resolver: zodResolver(loginSchema),
    })

    const login = useAuthStore((state) => state.login)

    const onSubmit = async (data: LoginFormData) => {
        // console.log(data);
        try {
            setLoading(true)
            await login(data.username, data.password)
            navigate("/chat", {replace: true})
            // console.log("Login successful");
        } catch(error) {
            console.error("Login failed:", error)
        } finally {
            setLoading(false);
        }
    }

    return (
        <div className="flex min-h-screen items-center justify-center px-4">
            <div className="w-full max-w-md">
                <div className="mb-8 text-center">
                    <h1 className="text-3xl font-bold">
                        Welcome back
                    </h1>

                    <p className="mt-2 text-sm text-gray-500">
                        Sign in continue to your account
                    </p>
                </div>

                <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
                    <div>
                        <label htmlFor="username" className="mb-2 block text-sm font-medium">Username</label>
                        <input type="username" id="username" {...register("username")} placeholder="Enter your username" className="w-full rounded-lg border px-4 py-3 outline-none focus:ring-2" />

                        {errors.username && (
                            <p className="mt-1 text-sm text-red-500">
                                {errors.username.message}
                            </p>
                        )}
                    </div>

                    <div>
                        <div className="mb-2 flex items-center justify-between">
                            <label htmlFor="password" className="text-sm font-medium">Password</label>
                            <Link to="/forgot-password" className="text-sm">Forgot password?</Link>
                        </div>

                        <input type="password" id="password" {...register("password")} placeholder="Enter your password" className="w-full rounded-lg border px-4 py-3 outline-none focus:ring-2" />

                        {errors.password && (
                            <p className="mt-1 text-sm text-red-500">
                                {errors.password.message}
                            </p>
                        )}
                    </div>

                    <button type="submit" disabled={loading} className="w-full rounded-lg px-4 py-3 font-medium">{loading ? "logining..." : "login"}</button>
                </form>

                <p className="mt-6 text-center text-sm">
                    Don't have an account?{" "}
                    <Link to="/register" className="font-medium">Create account</Link>
                </p>
            </div>
        </div>
    )
}