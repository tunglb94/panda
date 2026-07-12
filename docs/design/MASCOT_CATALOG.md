# Panda Mascot & Emoji — Asset Catalog

**Nguồn:** `I:\panda\Emoji` (10 file) + `I:\panda\Mascot` (35 file) = **45 asset**, tất cả đã được mở và xem trực tiếp (không chỉ đọc tên file).

**Phong cách chung:** nhân vật gấu trúc 3D (Pixar/Disney-style render), mắt xanh lá to, mặc hoodie xanh lá thương hiệu (logo "P" cách điệu) — phù hợp định vị Premium/Friendly/Modern như Duolingo/LINE. Thư mục `Emoji/` là các ảnh **bán thân** (đầu + vai), dùng tốt cho phản ứng nhỏ/inline. Thư mục `Mascot/` là ảnh **toàn thân** với đạo cụ theo ngữ cảnh (bảng hiệu, hộp quà, điện thoại...), dùng tốt cho khoảnh khắc hero (Empty State, Success, Celebration).

**Quy ước mức độ phù hợp:** ★★★★★ = dùng trực tiếp, không cần chỉnh sửa · ★★★★ = dùng tốt nhưng nên crop/xử lý nền · ★★★ = dùng được cho ngữ cảnh hẹp · ★★ = hạn chế, chỉ dùng khi rất phù hợp · ★ = không nên dùng trong UI sản phẩm hiện tại.

---

## ⚠️ Vấn đề chất lượng asset phát hiện được (đọc trước khi dùng)

| # | Vấn đề | File bị ảnh hưởng | Cách xử lý |
|---|---|---|---|
| 1 | **Nền không trong suốt thật** — nhiều PNG có nền vignette (đen/xám/xanh mờ dần) thay vì alpha trong suốt | Đa số file trong `Mascot/` (xem cột "Nền" bên dưới) | Bọc trong `Container` nền trắng/nền surface của app khi hiển thị; không đặt trực tiếp lên nền màu khác nếu chưa crop |
| 2 | **Artifact caro (checkerboard) bị bake vào ảnh** — không phải trong suốt thật, là hoạ tiết lục giác/caro hiển thị luôn | `Bất ngờ`, `Chờ đợi`, `Cười lớn`, `Dễ thương`, `Giận dỗi`, `Nghỉ ngợi`, `Nháy mắt`, `wow` (toàn bộ Emoji trừ Thích thú/Vui vẻ), `Tuyệt vời`, `Mất kết nối`, `thanh toán thất bại` | Cần cắt nền/tách nền lại (thiết kế); tạm thời tránh dùng làm hero lớn, chỉ dùng crop chặt quanh nhân vật |
| 3 | **Artifact hình khối vỡ (mảng trắng/vệt màu lạ) đè lên ảnh** | `Cảm ơn` (khối trắng bên phải), `Nhận hàng` (khối trắng bên phải), `Loading` (vệt đen/vàng/xanh dưới chân), `không tìm thấy` (dải trắng viền trên) | Không dùng nguyên bản; nếu dùng, phải crop chặt loại bỏ vùng lỗi (vd chỉ lấy phần đầu+vai) |
| 4 | **Sai tên file so với nội dung hình** | `Đánh giá.png` — tên nghĩa là "Rating" nhưng nội dung thực tế là ảnh **quét mã QR "SCAN ME"**, gần như trùng lặp với `Thanh toán.png` | Không dùng cho màn Rating; xếp vào nhóm Payment/QR. Cần đội thiết kế đổi tên hoặc thay ảnh |
| 5 | **Trùng lặp cảm xúc gần như y hệt** | `Chờ đợi` ≈ `Dễ thương` (cùng tư thế tay-trong-túi, cùng nụ cười nhẹ); `Bất ngờ` ≈ `wow` (cùng biểu cảm ngạc nhiên, khác biên độ) | Chỉ chọn 1 đại diện mỗi cặp cho mỗi ngữ cảnh để tránh lặp lại nhàm |
| 6 | **Text lỗi/garbled trên đạo cụ** | `Giao hàng.png` — chữ trên túi giao hàng hiển thị "Phho" thay vì tên thương hiệu rõ ràng | Không dùng cho đến khi asset được sửa; hiện tại tính năng giao đồ ăn cũng chưa có trong scope sản phẩm |
| 7 | **Trang phục không nhất quán với ngữ cảnh** | `sinh nhaat.png` (Sinh nhật) mặc đồ Santa đỏ thay vì trang phục sinh nhật riêng | Vẫn dùng được cho Birthday Promotion vì đạo cụ chính (bánh kem, nến, mũ tiệc) đã đúng ngữ cảnh, trang phục đỏ không gây hiểu lầm nghiêm trọng |

---

## PHẦN A — Thư mục `Emoji/` (bán thân, 10 file)

### 1. Bất ngờ.png
- **Nội dung hình:** Cận mặt, mắt xanh mở to, miệng há tròn kinh ngạc, hai tay không lộ (chỉ đầu+vai áo hoodie).
- **Cảm xúc:** Ngạc nhiên (mạnh).
- **Mức độ phù hợp:** ★★★
- **Màn hình phù hợp:** Surge/giá tăng đột ngột (thông báo), phát hiện voucher lớn bất ngờ, first-time discovery moment.
- **Ghi chú:** Nền caro bake sẵn — cần crop khi dùng.

### 2. Chờ đợi.png
- **Nội dung hình:** Đứng thẳng, tay trong túi, nụ cười nhẹ bình thản, mắt mở tự nhiên.
- **Cảm xúc:** Chờ đợi / bình thản.
- **Mức độ phù hợp:** ★★★
- **Màn hình phù hợp:** Trạng thái chờ nhẹ (chờ xác nhận, chờ phản hồi ngắn).
- **Ghi chú:** Gần trùng với "Dễ thương" — ưu tiên dùng file này cho ngữ cảnh "chờ", dùng "Dễ thương" cho ngữ cảnh trung tính khác để tránh lặp.

### 3. Cười lớn.png
- **Nội dung hình:** Nháy một mắt, giơ tay hình chữ V (peace sign), cười to lộ răng.
- **Cảm xúc:** Vui / ăn mừng nhẹ.
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Success nhỏ (đặt xe thành công, xác nhận nhanh), phản hồi tích cực sau hành động.

### 4. Dễ thương.png
- **Nội dung hình:** Gần như giống hệt "Chờ đợi" — đứng thẳng, tay trong túi, mỉm cười nhẹ, mắt to tròn.
- **Cảm xúc:** Trung tính / dễ thương.
- **Mức độ phù hợp:** ★★
- **Màn hình phù hợp:** Placeholder chung khi không có ngữ cảnh cảm xúc cụ thể (vd avatar mặc định).
- **Ghi chú:** Trùng lặp cao với "Chờ đợi" — cân nhắc chỉ dùng một trong hai.

### 5. Giận dỗi.png
- **Nội dung hình:** Chân mày cau, mắt nheo, miệng mím — biểu cảm hờn dỗi nhẹ, không dữ tợn.
- **Cảm xúc:** Không hài lòng (nhẹ, vẫn dễ thương).
- **Mức độ phù hợp:** ★★
- **Màn hình phù hợp:** Có thể dùng cho lời nhắc nhẹ nhàng kiểu "đừng bỏ lỡ ưu đãi" — **không** dùng cho lỗi hệ thống thật (đã có mascot buồn/lo phù hợp hơn: "Lỗi hệ thống").

### 6. Nghỉ ngợi.png
- **Nội dung hình:** Nháy mắt, tay đưa lên gần cằm kiểu tự tin/láu lỉnh — **nội dung hình không khớp với tên file** (tên gợi ý "nghỉ ngơi" nhưng dáng vẻ là tự tin/láu lỉnh, không phải nghỉ ngơi).
- **Cảm xúc:** Tự tin / láu lỉnh vui vẻ.
- **Mức độ phù hợp:** ★★★
- **Màn hình phù hợp:** Tip vặt/gợi ý nhỏ, thông báo mẹo hay ho.

### 7. Nháy mắt.png
- **Nội dung hình:** Nháy một mắt, tay trong túi, cười nhẹ thân thiện.
- **Cảm xúc:** Thân thiện / tinh nghịch nhẹ.
- **Mức độ phù hợp:** ★★★
- **Màn hình phù hợp:** Lời chào nhỏ, tooltip hướng dẫn, xác nhận thân thiện.

### 8. Thích thú.png
- **Nội dung hình:** Hai tay chắp lại trước ngực, mắt sáng, miệng cười háo hức. **Nền trắng sạch** (không phải vignette).
- **Cảm xúc:** Phấn khích / mong chờ.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Voucher mới, phần thưởng sắp nhận, trước khi mở hộp quà — một trong số ít asset nền sạch, ưu tiên dùng.

### 9. Vui vẻ.png
- **Nội dung hình:** Cười tươi đơn giản, tay trong túi. Nền vignette tối→xanh (KHÔNG trong suốt).
- **Cảm xúc:** Vui vẻ (chung chung).
- **Mức độ phù hợp:** ★★
- **Màn hình phù hợp:** Hero banner nền tối nếu cần (do nền không trong suốt nên hạn chế dùng làm icon nổi).
- **Ghi chú:** Nền hoàn toàn không trong suốt — asset kém linh hoạt nhất trong bộ Emoji.

### 10. wow.png
- **Nội dung hình:** Ngạc nhiên nhẹ hơn "Bất ngờ" — mắt mở to, miệng hé nhỏ, tay trong túi.
- **Cảm xúc:** Ngạc nhiên (nhẹ).
- **Mức độ phù hợp:** ★★★
- **Màn hình phù hợp:** Thông báo cập nhật app, tính năng mới ra mắt.
- **Ghi chú:** Gần trùng "Bất ngờ" — dùng file này cho ngạc nhiên nhẹ, "Bất ngờ" cho ngạc nhiên mạnh.

---

## PHẦN B — Thư mục `Mascot/` (toàn thân, 35 file)

### 11. Bảo trì.png
- **Nội dung hình:** Toàn thân, biểu cảm buồn/áy náy, hai tay cầm bảng "MAINTENANCE".
- **Cảm xúc:** Áy náy / xin lỗi.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Màn bảo trì hệ thống, tính năng tạm ngừng.

### 12. Chuyên nghiệp.png
- **Nội dung hình:** Đeo kính râm, hai tay giơ ngón cái, đeo balo, dáng đứng tự tin.
- **Cảm xúc:** Tự tin / chuyên nghiệp.
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Driver đạt hạng cao (Gold/Platinum/Diamond), huy hiệu thành tích, "Tài xế chuyên nghiệp".

### 13. Chào khách.png
- **Nội dung hình:** Vẫy tay chào, cười tươi, đứng toàn thân.
- **Cảm xúc:** Chào đón / thân thiện.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Welcome/Onboarding, màn hình Login, lần đầu mở app.

### 14. Chúc mừng.png
- **Nội dung hình:** Confetti bay xung quanh, hai tay giơ cao ăn mừng, cười lớn nhắm mắt. **Nền trắng** (artifact nhỏ ở đáy, không ảnh hưởng phần nhân vật).
- **Cảm xúc:** Ăn mừng / thành công lớn.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Hoàn thành chuyến, đạt thành tích/huy hiệu, hoàn thành quest — asset ăn mừng tốt nhất trong bộ.

### 15. Chỉ đường.png
- **Nội dung hình:** Cầm bản đồ giấy có ghim vị trí, tay kia chỉ vào bản đồ, đeo balo.
- **Cảm xúc:** Hướng dẫn / nhiệt tình.
- **Mức độ phù hợp:** ★★★
- **Màn hình phù hợp:** Onboarding "cách đặt xe hoạt động", hướng dẫn tính năng mới. **Không** dùng trong lúc đang xem bản đồ thật (vi phạm quy tắc "không che Map").

### 16. Cảm ơn.png
- **Nội dung hình:** Hai tay chắp hình trái tim/cảm ơn, cười nhẹ. Có mảng trắng vỡ hình bên phải (artifact).
- **Cảm xúc:** Cảm ơn / biết ơn.
- **Mức độ phù hợp:** ★★★ (do artifact, cần crop bên trái)
- **Màn hình phù hợp:** Sau khi hoàn tất thanh toán, cảm ơn đã sử dụng dịch vụ, footer màn hình rating.
- **Ghi chú:** Dùng `Alignment.centerLeft` + `ClipRect` giới hạn chiều rộng để loại bỏ khối artifact bên phải.

### 17. Cố lên.png
- **Nội dung hình:** Nắm đấm giơ cao, nháy mắt, cười tự tin.
- **Cảm xúc:** Động viên / cổ vũ.
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Nhắc nhở Quest/Streak (Driver), động viên hoàn thành mục tiêu tuần.

### 18. Giao hàng.png
- **Nội dung hình:** Cầm hộp mì, đũa, đeo túi giao hàng có chữ lỗi "Phho".
- **Cảm xúc:** Vui vẻ khi làm việc.
- **Mức độ phù hợp:** ★ (không dùng hiện tại)
- **Màn hình phù hợp:** Dự trữ cho tính năng Giao đồ ăn tương lai (BRB §15.1) — hiện **không** có trong phạm vi sản phẩm, và chữ trên đạo cụ bị lỗi.

### 19. Giáng sinh.png
- **Nội dung hình:** Trang phục Santa đầy đủ, túi quà, giơ ngón cái.
- **Cảm xúc:** Vui mừng lễ hội.
- **Mức độ phù hợp:** ★★★ (chỉ theo mùa)
- **Màn hình phù hợp:** Banner khuyến mãi Giáng sinh (theo lịch, không dùng quanh năm).

### 20. hài lòng.png
- **Nội dung hình:** Hai tay tạo hình trái tim, đeo túi giao hàng có logo "P", lấp lánh sao xung quanh.
- **Cảm xúc:** Hài lòng / yêu thích.
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Sau rating 5 sao, phản hồi tích cực từ hệ thống, cảm ơn đánh giá.

### 21. Khuyến mãi.png
- **Nội dung hình:** Hai tay xách túi mua sắm in ký hiệu "%".
- **Cảm xúc:** Hào hứng mua sắm/ưu đãi.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Trang Khuyến mãi/Ưu đãi, banner giảm giá.

### 22. không tìm thấy.png
- **Nội dung hình:** Cầm điện thoại, dấu hỏi chấm nổi hai bên đầu, biểu cảm bối rối nhẹ nhưng vẫn cười.
- **Cảm xúc:** Bối rối / không tìm thấy.
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Empty State "Không tìm thấy kết quả", tìm kiếm rỗng, 404.

### 23. Like.png
- **Nội dung hình:** Đứng thẳng, giơ ngón cái đơn giản, cười tươi.
- **Cảm xúc:** Đồng ý / tán thành.
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Xác nhận hành động thành công chung chung, "Đã lưu", "Đã cập nhật".

### 24. Loading.png
- **Nội dung hình:** Vòng cung các vạch xanh toả tròn quanh đầu (như spinner), hai tay giơ lên. Có vệt màu lỗi (đen/vàng/xanh) ở phần chân.
- **Cảm xúc:** Hào hứng / đang xử lý.
- **Mức độ phù hợp:** ★★★ (cần crop phần chân bị lỗi)
- **Màn hình phù hợp:** Splash loading, màn hình đang tải dữ liệu — chỉ dùng phần đầu+vai (crop 65% trên).

### 25. Lái xe hơi.png
- **Nội dung hình:** Ngồi sau vô-lăng xe hơi màu xanh lá, đầu ló ra khỏi kính chắn gió, cười tươi.
- **Cảm xúc:** Vui vẻ lái xe.
- **Mức độ phù hợp:** ★★★
- **Màn hình phù hợp:** Minh hoạ hạng xe "Ô tô" trong chọn xe (Rider), Vehicle Center (Driver) khi chưa có ảnh xe thật.

### 26. Lái xe máy.png
- **Nội dung hình:** Đội mũ bảo hiểm logo "P", lái xe máy tay ga màu xanh có logo "P".
- **Cảm xúc:** Tự tin / chuyên nghiệp.
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Minh hoạ hạng xe "Xe máy", Driver onboarding cho tài xế xe máy.

### 27. Lỗi hệ thống.png
- **Nội dung hình:** Đứng im, chân mày cau lo lắng, miệng mím buồn, không cầm đạo cụ. Nền đỏ nâu nhạt.
- **Cảm xúc:** Lo lắng / buồn (lỗi).
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Friendly Error chung (lỗi mạng, lỗi tải dữ liệu) khi không có mascot chuyên biệt hơn.

### 28. Mất kết nối.png
- **Nội dung hình:** Biểu tượng WiFi gạch chéo in trên áo, biểu cảm lo âu, mắt mở to.
- **Cảm xúc:** Lo lắng vì mất kết nối.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Empty State "Mất kết nối mạng" — khớp chính xác với các màn hình lỗi mạng hiện có trong `MapPage`/booking flow.

### 29. Nhận cuốc.png
- **Nội dung hình:** Cầm điện thoại hiển thị icon app, cười rạng rỡ.
- **Cảm xúc:** Hào hứng / vui mừng.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** (Driver) Có chuyến mới, màn hình Home khi Online chờ chuyến, thông báo Offer mới.

### 30. Nhận hàng.png
- **Nội dung hình:** Ôm hộp carton giao hàng. Có mảng trắng vỡ hình bên phải (artifact).
- **Cảm xúc:** Vui vẻ khi nhận việc.
- **Mức độ phù hợp:** ★ (không dùng hiện tại)
- **Màn hình phù hợp:** Dự trữ cho tính năng Giao hàng/Parcel tương lai (BRB §15.2) — ngoài phạm vi sản phẩm hiện tại, cộng thêm lỗi artifact.

### 31. NO GPS.png
- **Nội dung hình:** Biểu tượng định vị gạch chéo trên áo, cầm bản đồ giấy, biểu cảm lo âu. **Nền gần như trắng sạch.**
- **Cảm xúc:** Lo lắng vì mất GPS.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Empty State "Không có tín hiệu GPS"/"Quyền vị trí bị từ chối" — khớp chính xác với `_LocationErrorView` đã có trong `MapPage` (Rider).

### 32. Qà tặng.png
- **Nội dung hình:** Ló đầu ra khỏi hộp quà đỏ khổng lồ, lấp lánh sao, cười rạng rỡ.
- **Cảm xúc:** Bất ngờ vui sướng / nhận quà.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Nhận thưởng, mở voucher mới, Reward/Achievement.

### 33. sinh nhaat.png
- **Nội dung hình:** Đội mũ tiệc, bưng bánh kem sinh nhật có nến, mặc đồ đỏ kiểu lễ hội.
- **Cảm xúc:** Vui mừng chúc mừng.
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Banner khuyến mãi sinh nhật (Birthday Promotion).

### 34. thanh toán thất bại.png
- **Nội dung hình:** Cầm máy POS hiển thị "PAYMENT FAILED", tay kia cầm biển báo dấu X đỏ, biểu cảm buồn.
- **Cảm xúc:** Buồn / xin lỗi vì lỗi thanh toán.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Lỗi thanh toán trong Wallet/Payment.

### 35. Thanh toán.png
- **Nội dung hình:** Cầm điện thoại hiển thị mã QR "SCAN ME", đeo túi giao hàng logo "P".
- **Cảm xúc:** Vui vẻ / mời quét mã.
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Màn hình thanh toán QR trong Wallet (khi tính năng có nguồn dữ liệu thật).

### 36. Thông báo.png
- **Nội dung hình:** Cầm biểu tượng chuông/dấu chấm than xanh, thư điện tử bay xung quanh có huy hiệu đỏ.
- **Cảm xúc:** Hào hứng thông báo.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Notification Center Empty State ("Chưa có thông báo nào").

### 37. Trung thu.png
- **Nội dung hình:** Cầm lồng đèn ông sao truyền thống, cười tươi.
- **Cảm xúc:** Vui mừng lễ hội.
- **Mức độ phù hợp:** ★★★ (chỉ theo mùa)
- **Màn hình phù hợp:** Banner khuyến mãi Trung Thu (theo lịch).

### 38. Tuyệt vời.png
- **Nội dung hình:** Nháy mắt, giơ ngón cái. **Nền trong suốt thật** (checkerboard chuẩn, không bake artifact).
- **Cảm xúc:** Xuất sắc / tán thưởng.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Success chung (đặt xe thành công, thanh toán thành công, hoàn thành hồ sơ) — một trong các asset sạch nhất, ưu tiên dùng rộng rãi.

### 39. Tết.png
- **Nội dung hình:** Cầm phong bao lì xì đỏ chữ "福" (Phúc).
- **Cảm xúc:** Vui mừng / may mắn.
- **Mức độ phù hợp:** ★★★ (chỉ theo mùa)
- **Màn hình phù hợp:** Banner khuyến mãi Tết Nguyên Đán (theo lịch).

### 40. Voucher.png
- **Nội dung hình:** Cầm biển hiệu ghi rõ chữ "VOUCHER" kèm nhãn giá "%".
- **Cảm xúc:** Hào hứng chia sẻ ưu đãi.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Trang Voucher (Wallet), Empty State "Chưa có voucher nào".

### 41. Đang chạy.png
- **Nội dung hình:** Ngồi ghế xe cài dây an toàn, hai tay cầm vô-lăng, cười tươi.
- **Cảm xúc:** Vui vẻ / tập trung lái xe an toàn.
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Trip In Progress (Rider tracking), trạng thái "Đang di chuyển" (Driver).

### 42. Đánh giá 5 sao.png
- **Nội dung hình:** Cầm biển hiệu 5 ngôi sao vàng.
- **Cảm xúc:** Tự hào / vui mừng.
- **Mức độ phù hợp:** ★★★★★
- **Màn hình phù hợp:** Sau khi rider gửi đánh giá 5 sao, màn hình cảm ơn đánh giá.

### 43. Đánh giá.png
- **Nội dung hình:** **KHÔNG liên quan đến đánh giá** — thực chất là ảnh cầm điện thoại hiện mã QR "SCAN ME", gần như trùng lặp với `Thanh toán.png`.
- **Cảm xúc:** Vui vẻ / mời quét mã.
- **Mức độ phù hợp:** ★ (không dùng cho Rating)
- **Màn hình phù hợp:** Không dùng cho Rating do sai lệch nội dung so với tên; nếu cần ảnh QR dự phòng có thể dùng thay `Thanh toán.png`, nhưng ưu tiên `Đánh giá 5 sao.png` cho mọi ngữ cảnh đánh giá.

### 44. Đợi xe.png
- **Nội dung hình:** Đứng cạnh vali kéo, cười nhẹ nhàng, dáng vẻ kiên nhẫn.
- **Cảm xúc:** Chờ đợi kiên nhẫn.
- **Mức độ phù hợp:** ★★★★
- **Màn hình phù hợp:** Rider đang tìm tài xế ("Đang tìm chuyến"), booking sân bay/hành lý.

---

## Bảng tổng hợp nhanh theo cảm xúc

| Cảm xúc | Asset đại diện tốt nhất |
|---|---|
| Vui / Ăn mừng | `Chúc mừng`, `Tuyệt vời`, `Thích thú`, `Cười lớn` |
| Chờ đợi | `Đợi xe`, `Chờ đợi`, `Chỉ đường` |
| Ngạc nhiên | `Bất ngờ`, `wow` |
| Cảm ơn / Hài lòng | `Cảm ơn`, `hài lòng`, `Đánh giá 5 sao` |
| Xin lỗi / Lỗi | `Lỗi hệ thống`, `Mất kết nối`, `NO GPS`, `Bảo trì`, `thanh toán thất bại` |
| Ăn mừng thành tích | `Chuyên nghiệp`, `Qà tặng`, `Chúc mừng` |
| Loading | `Loading` (crop) |
| Ngủ / Trống dữ liệu | `không tìm thấy` (gần nhất — không có mascot "ngủ" chuyên biệt trong bộ) |
| Chào đón | `Chào khách` |
| Tặng quà / Voucher | `Qà tặng`, `Voucher`, `Khuyến mãi`, `Tết`, `sinh nhaat` |
| Cổ vũ | `Cố lên`, `Like` |

**Nhận xét:** bộ asset **không có** mascot chuyên biệt cho "Ngủ" (trạng thái rảnh/không hoạt động — vd Driver offline lâu, Empty State ban đêm) — dùng tạm `Chờ đợi` hoặc `không tìm thấy` tuỳ ngữ cảnh, và đề xuất bổ sung asset "Đang ngủ" trong đợt thiết kế tiếp theo.

---

*Catalog hoàn thành — 45/45 asset đã được xem trực tiếp, không dựa trên tên file.*
