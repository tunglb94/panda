# Panda — Full Business Validation: Business Audit

Mô phỏng hành vi (không phải load test) — mọi số liệu dưới đây tính trực tiếp từ dữ liệu mô phỏng thật, không suy đoán.

## 1. Revenue Leak

GMV lệch **0.000%** (0 VND) so với tổng đã phân bổ cho driver + platform + voucher + promotion. Không phát hiện thất thoát đáng kể.

## 2. Chuyến có Platform Profit < 0

**0** chuyến hoàn tất có lợi nhuận ước tính âm sau VAT + chi phí hạ tầng giả định (xem `unit_economics.json` cho giả định chi phí). Ví dụ: (không có)

## 3. Chuyến có Driver Income < 0

**0** chuyến có driver net âm. Ví dụ: (không có)

## 4. Voucher phát nhiều hơn dùng bao nhiêu %

Phát hành: **49**, đã dùng: **20** → **59.2%** voucher chưa từng được redeem trong lần chạy này.

## 5. Promotion ROI

- **manual_coupon**: ROI 9.00x, CPA 6,349 VND/lượt, 18 lượt redeem, repeat rate 94.4%

## 6. Surge làm Platform lỗ

**0** chuyến có surge (>1.0x) nhưng vẫn lỗ ước tính. Ví dụ: (không có)

## 7. Driver online 12h+

**0** tài xế từng online liên tục >=12h trong một ca (ngưỡng cứng của FatigueDecision).

## 8. Driver online nhưng 0 chuyến

**0** tài xế có online trong lần chạy này nhưng hoàn tất 0 chuyến.

## 9. Driver thu nhập cao bất thường

**0** tài xế có thu nhập/tuần vượt ngưỡng thống kê (mean + 3×độ lệch chuẩn).

## 10. Passenger spam voucher

**0** khách hàng redeem voucher >5 lần trong lần chạy này (ngưỡng thiết kế, xem ASSUMPTION).

## 11. Dispatch có bị thiên vị khu vực không

**ASSUMPTION quan trọng**: Dispatch thật (`RequestDispatchUseCase`/`offerNextDriver`) chỉ ghép theo khoảng cách gần nhất, không có khái niệm "ưu tiên khu vực" trong thuật toán. Chênh lệch accept-rate giữa các khu vực dưới đây phản ánh **phân bố cung tài xế thực tế theo khu vực**, không phải thiên vị thuật toán.

| Khu vực | Requested | Accept Rate | ETA TB | Cầu/Cung TB |
|---|---:|---:|---:|---:|
| Sân bay | 82 | 51.2% | 40.1 phút | 7.83x |
| Bến xe | 77 | 100.0% | 32.0 phút | 3.15x |
| Trung tâm (CBD) | 122 | 100.0% | 19.6 phút | 2.52x |
| Khu vui chơi | 79 | 100.0% | 27.5 phút | 1.88x |
| Bệnh viện | 70 | 100.0% | 21.9 phút | 2.84x |
| Khu công nghiệp (KCN) | 74 | 100.0% | 29.5 phút | 1.54x |
| Khu dân cư | 91 | 98.9% | 26.4 phút | 1.46x |
| Trường học | 71 | 100.0% | 26.1 phút | 2.32x |

## 12. ETA cao nhất

Khu vực **Sân bay** có ETA trung bình cao nhất: **40.1 phút**.

## 13. Khu vực thiếu driver

- Sân bay: cầu/cung trung bình **7.8x**
- Bến xe: cầu/cung trung bình **3.2x**
- Bệnh viện: cầu/cung trung bình **2.8x**

## 14. Khu vực thừa driver

- Khu dân cư: cầu/cung trung bình **1.46x**
- Khu công nghiệp (KCN): cầu/cung trung bình **1.54x**
- Khu vui chơi: cầu/cung trung bình **1.88x**

## 15. Ride vs Delivery Ratio

Ride: **461**, Delivery: **164** → tỉ lệ **2.81:1**.

## 16. Airport Profit

86 chuyến liên quan Sân bay, lợi nhuận trung bình/chuyến **17,892 VND**, tổng **1,538,785 VND**.

## 17. Peak Hour Profit

171 chuyến giờ cao điểm (07:00-09:00, 17:00-20:00), lợi nhuận trung bình/chuyến **10,883 VND**.

## 18. Off Peak Profit

454 chuyến giờ thấp điểm, lợi nhuận trung bình/chuyến **10,493 VND**.

## 19. Weather ảnh hưởng thế nào

- **heavy_rain**: 351 chuyến, fare TB 64,468 VND, surge TB 1.26x, cancel rate 0.0%
- **sunny**: 274 chuyến, fare TB 58,488 VND, surge TB 1.11x, cancel rate 0.7%

## 20. Top 50 Anomaly

Xem `top_50_anomalies.json` cho danh sách đầy đủ, xếp hạng theo mức độ nghiêm trọng.

---

## Validation (đối chiếu với validation_report.json)

**Passed: true**


## Bugs phát hiện (không tự sửa)

Không phát hiện bug mới nào trong lần audit này ngoài các bug đã được sửa ở phase trước (xem CHANGELOG).


## ASSUMPTION (logic không hợp lý/giả định, không tự sửa)

1. **VAT 10%** — BRB tự loại VAT khỏi phạm vi ("calculated and remitted by the Finance team independently", business-rule-bible-v1.0.md dòng 986) — 10% là mức chuẩn thuế Việt Nam, không phải số BRB.
2. **Chi phí Cloud/Map/SMS mỗi chuyến (400đ/250đ/350đ)** — Giả định thiết kế, không có dữ liệu tài chính thật của Panda để đối chiếu — xem stats/unit_economics.go.
3. **Motorcycle fare = 60% giá Car** — BRB không định nghĩa rate cho xe máy — giả định tỉ lệ điển hình thị trường Đông Nam Á, không phải số BRB — xem integration/pricing_adapter.go.
4. **"Bike Plus" không tồn tại trong production** — pricing_analytics.json báo cáo trung thực not_modeled=true thay vì bịa công thức giá — Panda chưa từng xây dựng hạng xe máy cao cấp.
5. **Dispatch "bias" phản ánh cung tài xế, không phải thuật toán** — offerNextDriver chỉ ghép theo khoảng cách gần nhất — chênh lệch accept-rate theo khu vực (§11) là artifact của phân bố cung, không phải lỗi thiên vị trong code Dispatch.
6. **Ngưỡng "voucher spam" (>5 lần/lần chạy)** — Ngưỡng thiết kế mô phỏng, không phải số liệu chống gian lận thật — không có dữ liệu lịch sử để hiệu chỉnh.
7. **Ngưỡng driver income outlier (mean + 3×stddev)** — Ngưỡng thống kê tiêu chuẩn, không phải chính sách BRB nào.
8. **Zone/thời tiết/traffic là hằng số thiết kế mô phỏng** — Toạ độ khu vực, base demand weight, xác suất thời tiết/traffic không đến từ dữ liệu thành phố thật — xem domain/entity/city.go, simulation/scenario_scheduler.go.
9. **Delivery fare dùng chung Pricing Engine với Ride** — Production chưa tích hợp Pricing cho Delivery (xác nhận qua backend/services/trip/app/complete_delivery.go's doc comment) — simulation ước tính fare Delivery bằng cùng công thức Ride, giống đúng tiền lệ Rider app's DeliveryFormPage đã làm.
10. **Peak Hour = 07:00-09:00 và 17:00-20:00** — Định nghĩa thiết kế mô phỏng dùng cho §17/§18 — không phải khung giờ BRB chính thức nào (BRB có Peak Hour Surcharge riêng nhưng không định nghĩa khung giờ cố định theo cách này).
11. **Hai "Acceptance Rate" khác nhau dùng cùng một tên** — driver_analytics.json's acceptance_rate_percent đo tỉ lệ MỘT tài xế đồng ý khi được đề nghị một chuyến cụ thể (offersAccepted/(offersAccepted+offersRejected), driverAcceptsOffer trong ride_flow.go); CEO_report.html/executive_dashboard.html's "Acceptance Rate" đo tỉ lệ MỘT yêu cầu cuối cùng có tài xế nhận (completed/requested, sau khi đã thử nhiều tài xế qua vòng lặp resolveOffer). Cả hai đều tính đúng theo định nghĩa riêng, nhưng tên gọi giống nhau dễ gây nhầm lẫn khi đối chiếu hai file — quan sát được trong lần chạy full-scale: 81.4% (per-offer) vs 99.8% (per-request).
12. **Demand/Supply ratio ở scale 1000/10000 không nên đọc như tỉ lệ dân số** — Ở cấu hình 1000 tài xế/10.000 khách (tỉ lệ dân số 1:10), heatmap ghi nhận cầu/cung TỨC THỜI (theo tick) ở mức đồng đều 42-44x tại MỌI khu vực — kể cả các khu vực được xếp "thừa driver" trong §14 (so sánh tương đối, không phải oversupply thật, ratio luôn >>1). Vẫn accept rate 99.8% nhờ cơ chế retry nhiều tài xế gần nhất, không phải vì có đủ cung tại một thời điểm — cần đọc con số 42-44x như "áp lực cầu tức thời cao", không phải "khách phải chờ 42x lâu hơn bình thường".
