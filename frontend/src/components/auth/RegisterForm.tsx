import { Link, useNavigate } from "react-router";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import axios from "axios";
import { registerSchema, type RegisterFormData } from "../../types/auth";
import { useAuthStore } from "../../stores/auth.store";
import { useState } from "react";

export default function RegisterForm() {
    const [loading, setLoading] = useState(false)
    const navigate = useNavigate()
    const {register, handleSubmit, formState: { errors }} = useForm<RegisterFormData>({
        resolver: zodResolver(registerSchema)
    })

    const registers = useAuthStore(state => state.register);

    const onSubmit = async (data: RegisterFormData) => {
        console.log("error");
        try {
            setLoading(true);
            await registers(
                data.username, 
                data.password
            );
            navigate("/login", {replace: true})
            console.log("đăng ký tài khoản thành công");
        } catch (error) {
            if(axios.isAxiosError(error)) {
                alert(error.response?.data.error);
                return;
            }
            alert("có lỗi xảy ra!");
        } finally {
            setLoading(false);
        }
    }

    return (
        <div className="flex min-h-screen items-center justify-center px-4">
            <div className="w-full max-w-md">
                <div className="mb-8 text-center">
                    <h1 className="text-3xl font-bold">
                        Create account 
                    </h1>

                    <p className="mt-2 text-sm text-gray-500">
                        Create your account to get started
                    </p>
                </div>

                <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
                    <div>
                        <label htmlFor="username" className="mb-2 block text-sm font-medium">Username</label>
                        <input type="text" id="username" {...register("username")} placeholder="Enter your username" className="w-full rounded-lg border px-4 py-3 outline-none focus:ring-2" />

                        {errors.username && (
                            <p className="mt-1 text-sm text-red-500">
                                {errors.username.message}
                            </p>
                        )}
                    </div>

                    <div>
                        <label htmlFor="password" className="mb-2 block text-sm font-medium">Password</label>
                        <input type="password" id="password" {...register("password")} placeholder="Create a password" className="w-full rounded-lg border px-4 py-3 outline-none focus:ring-2" />

                        {errors.password && (
                            <p className="mt-1 text-sm text-red-500">
                                {errors.password.message}
                            </p>
                        )}
                    </div>

                    <button type="submit" disabled={loading} className="w-full rounded-lg px-4 py-3 font-medium">
                        {loading ? "Creating account..." : "Create account"}
                    </button>
                </form>

                <p className="mt-6 text-center text-sm">
                    Already have an account?{" "}
                    <Link to="/login" className="font-medium">
                        Sign in 
                    </Link>
                </p>
            </div>
        </div>
    )
}