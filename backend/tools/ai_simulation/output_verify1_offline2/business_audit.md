# Panda — Full Business Validation: Business Audit

Mô phỏng hành vi (không phải load test) — mọi số liệu dưới đây tính trực tiếp từ dữ liệu mô phỏng thật, không suy đoán.

## 1. Revenue Leak

GMV lệch **0.000%** (0 VND) so với tổng đã phân bổ cho driver + platform + voucher + promotion. Không phát hiện thất thoát đáng kể.

## 2. Chuyến có Platform Profit < 0

**0** chuyến hoàn tất có lợi nhuận ước tính âm sau VAT + chi phí hạ tầng giả định (xem `unit_economics.json` cho giả định chi phí). Ví dụ: (không có)

## 3. Chuyến có Driver Income < 0

**0** chuyến có driver net âm. Ví dụ: (không có)

## 4. Voucher phát nhiều hơn dùng bao nhiêu %

Phát hành: **110**, đã dùng: **68** → **38.2%** voucher chưa từng được redeem trong lần chạy này.

## 5. Promotion ROI

- **manual_coupon**: ROI 9.00x, CPA 5,815 VND/lượt, 55 lượt redeem, repeat rate 83.6%
- **first_ride**: ROI 1.00x, CPA 19,156 VND/lượt, 1 lượt redeem, repeat rate 100.0%

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
| Sân bay | 155 | 52.3% | 41.2 phút | 1.33x |
| Bến xe | 188 | 88.8% | 31.2 phút | 3.14x |
| Trung tâm (CBD) | 271 | 97.8% | 20.9 phút | 4.65x |
| Khu vui chơi | 187 | 92.5% | 28.9 phút | 2.20x |
| Bệnh viện | 146 | 94.5% | 25.0 phút | 6.20x |
| Khu công nghiệp (KCN) | 170 | 82.4% | 34.1 phút | 1.64x |
| Khu dân cư | 208 | 88.0% | 30.2 phút | 2.51x |
| Trường học | 207 | 93.7% | 26.8 phút | 4.48x |

## 12. ETA cao nhất

Khu vực **Sân bay** có ETA trung bình cao nhất: **41.2 phút**.

## 13. Khu vực thiếu driver

- Bệnh viện: cầu/cung trung bình **6.2x**
- Trung tâm (CBD): cầu/cung trung bình **4.7x**
- Trường học: cầu/cung trung bình **4.5x**

## 14. Khu vực thừa driver

- Sân bay: cầu/cung trung bình **1.33x**
- Khu công nghiệp (KCN): cầu/cung trung bình **1.64x**
- Khu vui chơi: cầu/cung trung bình **2.20x**

## 15. Ride vs Delivery Ratio

Ride: **1080**, Delivery: **261** → tỉ lệ **4.14:1**.

## 16. Airport Profit

172 chuyến liên quan Sân bay, lợi nhuận trung bình/chuyến **16,213 VND**, tổng **2,788,653 VND**.

## 17. Peak Hour Profit

420 chuyến giờ cao điểm (07:00-09:00, 17:00-20:00), lợi nhuận trung bình/chuyến **9,640 VND**.

## 18. Off Peak Profit

921 chuyến giờ thấp điểm, lợi nhuận trung bình/chuyến **9,053 VND**.

## 19. Weather ảnh hưởng thế nào

- **sunny**: 1341 chuyến, fare TB 53,573 VND, surge TB 1.11x, cancel rate 0.6%

## 20. Top 50 Anomaly

Xem `top_50_anomalies.json` cho danh sách đầy đủ, xếp hạng theo mức độ nghiêm trọng.

---

## Validation (đối chiếu với validation_report.json)

**Passed: true**


## Bugs phát hiện (không tự sửa)

### Bug 1: --seed không tạo ra kết quả xác định (deterministic) giữa các lần chạy

- **Nguyên nhân:** World.Drivers/World.Riders là map[string]*entity.DriverAgent/RiderAgent — Go cố ý ngẫu nhiên hoá thứ tự duyệt map giữa mỗi lần chạy chương trình. Mọi vòng lặp `for _, d := range w.Drivers` / `for _, r := range w.Riders` (simulation/engine.go's processTick/evaluateDriverState, onNewDay, RefreshZoneCounters...) tiêu thụ số ngẫu nhiên từ *rand.Rand đã seed theo đúng thứ tự đó — thứ tự duyệt khác nhau → chuỗi số ngẫu nhiên được rút ra khác nhau → kết quả khác nhau, dù --seed giống hệt nhau.
- **Ảnh hưởng:** Đã tự kiểm chứng trực tiếp: chạy `--seed=777` hai lần với cấu hình giống hệt nhau (30 driver/200 rider/2 ngày) cho ra 666 vs 586 requested (~12% lệch), GMV lệch ~20%. Điều này có nghĩa: (1) không có lần chạy nào trong báo cáo audit này có thể tái hiện chính xác byte-for-byte chỉ bằng cách chạy lại cùng --seed; (2) --seed hiện tại chỉ hữu ích để tạo dữ liệu ngẫu nhiên có kiểm soát ở mức phân phối thống kê, không phải để debug/regression-test một kết quả cụ thể. Không ảnh hưởng đến tính hợp lệ của các kết luận kinh doanh (vẫn là dữ liệu thật từ một lần chạy thật), nhưng ảnh hưởng đến khả năng tái lập số liệu chính xác.
- **Cách tái hiện:** go run ./backend/tools/ai_simulation --drivers=30 --riders=200 --days=2 --seed=777 --out=/tmp/a && go run ./backend/tools/ai_simulation --drivers=30 --riders=200 --days=2 --seed=777 --out=/tmp/b && diff /tmp/a/simulation_report.json /tmp/b/simulation_report.json
- **File liên quan:** `backend/tools/ai_simulation/simulation/world.go (Drivers/Riders map fields), engine.go (mọi range trên 2 map này), seed.go`


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
