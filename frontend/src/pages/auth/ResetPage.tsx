import ResetPasswordForm from "../../components/auth/ResetForm";

export default function ResetPasswordPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-100 px-4">
      <div className="w-full max-w-md rounded-xl bg-white p-8 shadow-lg">
        <h1 className="mb-2 text-center text-3xl font-bold">
          Reset Password
        </h1>

        <p className="mb-8 text-center text-gray-500">
          Nhập username, mã OTP và mật khẩu mới.
        </p>

        <ResetPasswordForm />
      </div>
    </div>
  );
}