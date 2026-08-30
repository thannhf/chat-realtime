# Chat Realtime Application

Một dự án cá nhân mã nguồn mở (Public Version) nhằm xây dựng hệ thống trò chuyện trực tuyến thời gian thực hiệu năng cao. Dự án được phát triển với mục tiêu học tập, tối ưu hóa tư duy kiến trúc và làm sản phẩm điểm nhấn trong hồ sơ năng lực (CV) của mình.

Dự án sẽ liên tục được tối ưu hóa, nâng cấp cấu trúc và bổ sung tính năng mới từ nay cho đến **Mùa hè năm 2027**.

---

## Trạng thái hiện tại của phiên bản Public

Đây là phiên bản rút gọn (Public), tập trung vào các tính năng lõi để cộng đồng dễ dàng tiếp cận, chạy thử và đóng góp ý kiến. Để giữ dự án gọn nhẹ trong giai đoạn này, các luồng xử lý phức tạp (Complex Flows) và tài liệu chi tiết (Docs) tạm thời được lược bỏ. Mình sẽ cập nhật đầy đủ tài liệu sau.

---

## Công nghệ sử dụng (Tech Stack)

Dự án được xây dựng dựa trên những công nghệ và thư viện mới nhất hiện nay:

### 🖥️ Frontend (React 19 & Vite 8)
- **Core:** React 19, TypeScript 6, Vite 8 (Hiệu năng compile vượt trội).
- **State Management:** `zustand` (Quản lý trạng thái gọn nhẹ, tối ưu).
- **Data Fetching:** `@tanstack/react-query v5` (Quản lý server state, caching mạnh mẽ).
- **Styling:** Tailwind CSS v4 (`@tailwindcss/vite`) - Phiên bản mới nhất với hiệu năng tối ưu bằng Rust.
- **Form & Validation:** `react-hook-form`, `@hookform/resolvers`, `zod`.
- **Routing:** `react-router v8`.
- **HTTP Client:** `axios`.
- **Icons & Icons:** `lucide-react`, `date-fns`.

### Backend & Database
- **Language:** **Golang** (Đảm bảo tốc độ xử lý vượt trội và tối ưu tài nguyên hệ thống cho tác vụ realtime).
- **Database:** **PostgreSQL** (Lưu trữ dữ liệu tin nhắn, người dùng bền vững).
- **Caching:** **Redis** (Tăng tốc độ truy vấn, quản lý session và hỗ trợ cơ chế realtime).

---

## Khuyến khích đóng góp & Góp ý (Contributing)

Vì đây là sản phẩm tâm huyết phục vụ cho mục đích phát triển kỹ năng chuyên môn, **mọi ý kiến đóng góp, nhận xét về cấu trúc code hay kiến trúc hệ thống từ bạn đều vô cùng quý giá**.

Nếu bạn ghé thăm và muốn giúp dự án hoàn thiện hơn:

1. **Góp ý trực tiếp:** Bạn có thể mở một [Issue](https://github.com) để nhận xét về cấu trúc thư mục, cách viết Golang/React hoặc gợi ý thêm tính năng.
2. **Đóng góp mã nguồn:** 
   - Fork dự án về kho lưu trữ của bạn.
   - Tạo một nhánh mới (`git checkout -b feature/AmazingFeature`).
   - Commit thay đổi (`git commit -m 'Add some AmazingFeature'`).
   - Push lên nhánh đó (`git push origin feature/AmazingFeature`).
   - Mở một **Pull Request** để mình review và gộp code.

---

## Tài liệu kỹ thuật (Documentation)

*Lưu ý: Hệ thống sơ đồ luồng (flowchart) chi tiết và tài liệu hướng dẫn cài đặt môi trường cục bộ (Local Setup) hiện đang được biên soạn. Mình sẽ cập nhật đầy đủ trong thư mục `/documents` ở các phiên bản kế tiếp. Cảm ơn sự quan tâm của bạn!*
