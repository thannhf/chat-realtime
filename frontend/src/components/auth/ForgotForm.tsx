import { Link, useNavigate } from "react-router";
import { forgotPasswordSchema, type ForgotPasswordFormData } from "../../types/auth";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { useAuthStore } from "../../stores/auth.store";

export default function ForgotPasswordForm() {
    const navigate = useNavigate()
    const {register, handleSubmit, formState: { errors }} = useForm<ForgotPasswordFormData>({
        resolver: zodResolver(forgotPasswordSchema)
    });

    const forgotPassword = useAuthStore(state => state.forgotPassword);

    const onSubmit = async (data: ForgotPasswordFormData) => {
        try {
            await forgotPassword(data.username);
            alert("OTP đã được gửi");
            navigate("/reset-password", {
                state: {
                    username: data.username,
                },
            });
        } catch (error) {
            console.error(error);
            if (axios.isAxiosError(error)) {
                console.log(error.response);
                console.log(error.response?.data);
                return;
            }

            console.log(error);
        }
    }

    return (
        <div className="flex min-h-screen items-center justify-center px-4">
            <div className="w-full max-w-md">
                <div className="mb-8 text-center">
                    <h1 className="text-3xl font-bold">
                        Forgot password?
                    </h1>

                    <p className="mt-2 text-sm text-gray-500">
                        Enter your username and we'll send you a reset link. 
                    </p>
                </div>

                <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
                    <div>
                        <label htmlFor="username" className="mb-2 block text-sm font-medium">Username</label>
                        <input type="username" id="username" {...register("username")} className="w-full rounded-lg border px-4 py-3 outline-none focus:ring-2" placeholder="Enter your username" />

                        {errors.username && (
                            <p className="mt-1 text-sm text-red-500">
                                {errors.username.message}
                            </p>
                        )}
                    </div>

                    <button type="submit" className="w-full rounded-lg px-4 py-3 font-medium">Send reset link</button>
                </form>

                <p className="mt-6 text-center text sm">
                    Remember your password?{" "}
                    <Link to="/login" className="font-medium">
                        Back to login 
                    </Link>
                </p>
            </div>
        </div>
    )
}