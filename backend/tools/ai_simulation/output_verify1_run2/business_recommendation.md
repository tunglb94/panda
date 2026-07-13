# Panda Simulation — Đề xuất cải thiện kinh doanh

_Lưu ý: chỉ 13 đề xuất đủ điều kiện nổi bật trong lần chạy này (ngưỡng tối đa là 30)._

## 1. Giảm cancellation rate (hiện tại 25.3% tổng yêu cầu) bằng cách cải thiện độ chính xác ETA hiển thị cho khách trước khi đặt.

- **Priority:** High
- **Expected Impact:** Giảm cancellation rate, tăng trải nghiệm khách hàng
- **Risk:** Cần xác định nguyên nhân gốc trước khi hành động

## 2. Điều tra nguyên nhân vận hành tại Trường học — khu vực có số chuyến huỷ sau khi đã ghép tài xế cao nhất (70 chuyến).

- **Priority:** High
- **Expected Impact:** Giảm cancellation rate
- **Risk:** Cần điều tra vận hành cụ thể trước khi hành động

## 3. Ưu tiên phân bổ tài xế và khuyến mãi mục tiêu tại Trung tâm (CBD) — khu vực có nhu cầu cao nhất (244 yêu cầu).

- **Priority:** Medium
- **Expected Impact:** Tối đa hoá GMV tại khu vực có nhu cầu cao nhất
- **Risk:** Có thể gây mất cân bằng phân bổ tài xế ở khu vực khác

## 4. Tăng incentive/thưởng theo giờ cho tài xế hoạt động tại Bệnh viện — tỉ lệ cầu/cung trung bình đang ở mức 8.6x.

- **Priority:** High
- **Expected Impact:** Giảm ETA và tăng acceptance rate tại khu vực thiếu cung
- **Risk:** Chi phí incentive tăng ngắn hạn

## 5. Mở rộng ngân sách cho "Manual Coupon" — ROI cao nhất (9.0x), 32 lượt redeem.

- **Priority:** Medium
- **Expected Impact:** Tăng GMV và số chuyến với chi phí khuyến mãi hiệu quả
- **Risk:** Ngân sách marketing tăng

## 6. Tăng mật độ tài xế tại các khu vực/khung giờ có ETA cao — ETA trung bình hiện tại 31.1 phút.

- **Priority:** Medium
- **Expected Impact:** Giảm ETA trung bình, tăng trải nghiệm khách hàng
- **Risk:** Cần tăng mật độ tài xế, có thể tăng chi phí incentive

## 7. Xem xét lộ trình lên hạng (tier) rõ ràng hơn cho tài xế mới — 44.0% tài xế vẫn ở hạng Bronze (hoa hồng 20%, cao nhất theo BRB §7.1).

- **Priority:** Medium
- **Expected Impact:** Tăng driver income trung bình, giảm rủi ro nghỉ việc ở nhóm mới
- **Risk:** Giảm doanh thu hoa hồng ngắn hạn nếu hạ ngưỡng lên hạng

## 8. Xem xét giảm ngân sách hoặc thắt chặt điều kiện áp dụng cho "First Ride" — ROI thấp nhất trong các chương trình đã redeem (2.3x).

- **Priority:** Medium
- **Expected Impact:** Giảm chi phí voucher/promotion, tăng lợi nhuận ròng
- **Risk:** Có thể làm giảm số chuyến nếu cắt giảm quá mạnh

## 9. Thiết kế ưu đãi nâng hạng thành viên — 68.6% khách hàng vẫn ở hạng Free.

- **Priority:** Medium
- **Expected Impact:** Tăng retention và giảm price sensitivity trung bình
- **Risk:** Cần thiết kế ưu đãi hấp dẫn nhưng không ăn mòn margin

## 10. Điều tra nguyên nhân tài xế từ chối chuyến (fare thấp, khoảng cách xa, khu vực kém an toàn) — acceptance rate hiện tại 61.7%.

- **Priority:** High
- **Expected Impact:** Tăng acceptance rate, giảm thời gian tìm tài xế cho khách
- **Risk:** Cần điều tra fare/khoảng cách cụ thể

## 11. Rà soát cấu trúc giá cho hạng "Delivery Bike" — lợi nhuận trung bình/chuyến thấp hơn đáng kể so với "Ride Car XL".

- **Priority:** Low
- **Expected Impact:** Cải thiện lợi nhuận ở hạng xe yếu nhất
- **Risk:** Cần phân tích thêm nguyên nhân chênh lệch trước khi đổi giá

## 12. Đơn giản hoá điều kiện áp dụng voucher — chỉ 29.7% lượt được đề nghị chọn dùng ngay.

- **Priority:** Low
- **Expected Impact:** Tăng tỉ lệ redeem voucher đã phát hành
- **Risk:** Chi phí voucher có thể tăng nếu quá dễ áp dụng

## 13. Theo dõi biến động GMV theo Ride (41,920,977 VND) và Delivery (10,282,814 VND) theo thời gian thực để phát hiện tính mùa vụ sớm.

- **Priority:** Low
- **Expected Impact:** Theo dõi tính mùa vụ để tối ưu phân bổ ngân sách marketing
- **Risk:** Không có

---
_Mỗi đề xuất được sinh ra từ một điều kiện thật đo được trong dữ liệu mô phỏng (xem simulation_summary.md và các file *_analytics.json/*.json liên quan); Priority/Expected Impact/Risk là nhận định nghiệp vụ đi kèm, không phải số liệu đo được._
