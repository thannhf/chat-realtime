import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useLocation, useNavigate } from "react-router";
import { z } from "zod";

import { useAuthStore } from "../../stores/auth.store";

const resetPasswordSchema = z
  .object({
    username: z
      .string()
      .min(3, "Username phải có ít nhất 3 ký tự"),

    otp: z
      .string()
      .length(6, "OTP phải gồm đúng 6 ký tự"),

    newPassword: z
      .string()
      .min(8, "Mật khẩu phải có ít nhất 8 ký tự"),

    confirmPassword: z.string(),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    path: ["confirmPassword"],
    message: "Mật khẩu xác nhận không khớp",
  });

type ResetPasswordFormData = z.infer<typeof resetPasswordSchema>;

export default function ResetPasswordForm() {
  const navigate = useNavigate();
  const location = useLocation();

  const resetPassword = useAuthStore(
    (state) => state.resetPassword,
  );

  const [loading, setLoading] = useState(false);

  const username = location.state?.username ?? "";

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ResetPasswordFormData>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: {
      username,
    },
  });

  const onSubmit = async (
    data: ResetPasswordFormData,
  ) => {
    try {
      setLoading(true);

      await resetPassword(
        data.username,
        data.otp,
        data.newPassword,
      );

      alert("Đặt lại mật khẩu thành công!");

      navigate("/login", {
        replace: true,
      });
    } catch (error) {
      if (axios.isAxiosError(error)) {
        alert(
          error.response?.data?.error ??
            "Có lỗi xảy ra",
        );
      } else {
        alert("Có lỗi xảy ra");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="space-y-5"
    >
      <div>
        <label className="mb-2 block text-sm font-medium">
          Username
        </label>

        <input
          {...register("username")}
          className="w-full rounded-lg border p-3"
          placeholder="Nhập username"
        />

        {errors.username && (
          <p className="mt-1 text-sm text-red-500">
            {errors.username.message}
          </p>
        )}
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium">
          OTP
        </label>

        <input
          {...register("otp")}
          className="w-full rounded-lg border p-3"
          placeholder="Nhập mã OTP"
        />

        {errors.otp && (
          <p className="mt-1 text-sm text-red-500">
            {errors.otp.message}
          </p>
        )}
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium">
          New Password
        </label>

        <input
          type="password"
          {...register("newPassword")}
          className="w-full rounded-lg border p-3"
          placeholder="Nhập mật khẩu mới"
        />

        {errors.newPassword && (
          <p className="mt-1 text-sm text-red-500">
            {errors.newPassword.message}
          </p>
        )}
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium">
          Confirm Password
        </label>

        <input
          type="password"
          {...register("confirmPassword")}
          className="w-full rounded-lg border p-3"
          placeholder="Nhập lại mật khẩu"
        />

        {errors.confirmPassword && (
          <p className="mt-1 text-sm text-red-500">
            {errors.confirmPassword.message}
          </p>
        )}
      </div>

      <button
        type="submit"
        disabled={loading}
        className="w-full rounded-lg bg-blue-600 py-3 text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {loading
          ? "Resetting..."
          : "Reset Password"}
      </button>

      <p className="text-center text-sm">
        Quay lại{" "}
        <Link
          to="/login"
          className="text-blue-600 hover:underline"
        >
          Đăng nhập
        </Link>
      </p>
    </form>
  );
}