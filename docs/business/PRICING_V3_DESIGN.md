# Panda — Pricing V3: Kiến trúc Kinh tế Toàn diện (Design)

**Document Classification:** Internal — Confidential
**Authority:** CPO · CFO · CTO (phê duyệt bắt buộc trước khi bất kỳ phần nào được triển khai)
**Effective Date:** 2026-07-11
**Status:** THIẾT KẾ KIẾN TRÚC — không phải task code, không sửa bất kỳ source code nào (BRB, Pricing Service, Promotion Engine, UI đều giữ nguyên), không build, không commit.
**Nguồn sự thật khi có mâu thuẫn:** `docs/business/business-rule-bible-v1.0.md` (BRB) vẫn là SSOT vận hành hôm nay. Tài liệu này là **thiết kế cho phiên bản kế tiếp (V3)** — mọi con số/cấu trúc mới đều là **đề xuất chờ duyệt**, không có hiệu lực cho đến khi đi qua quy trình tu chính chính thức (Constitution Article XI).
**Tài liệu đã đọc trước khi viết:** BRB v1.0 (Part 2 Pricing, Part 3-4 Promotion/Voucher, Part 7-9 Driver Economy/Incentive/Performance), `PRICING_STRATEGY.md`, `ECONOMY_ENGINE.md`, `MARKET_PRICING_RESEARCH.md` (tài liệu này **kế thừa trực tiếp** — mọi số liệu thị trường, 312-kịch-bản, driver/platform economics đã tính ở đó được **tái sử dụng**, không tính lại từ đầu), `backend/services/pricing/domain/entity/fare.go` + `backend/services/pricing/simulation/*.go` (production + simulation engine), `backend/services/promotion/domain/entity/voucher.go` + `promotion_type.go` + `app/promotion_service.go` (engine thật, single-voucher-wins model).

---

## TÓM TẮT ĐIỀU HÀNH

`MARKET_PRICING_RESEARCH.md` đã xác nhận bằng số liệu thật: Panda rẻ hơn thị trường 42.5% trung bình, nguyên nhân là **/km quá thấp** (không phải Base Fare), quãng đường dài gần như không có lợi nhuận tài xế (chi phí xăng+khấu hao ăn hết phần thu theo km), và công thức phẳng (1 mức /km duy nhất) khiến bảng giá đề xuất tạm thời của tài liệu đó vi phạm trần giá tuyệt đối ở quãng ≥50km.

**Pricing V3 giải quyết cả bốn vấn đề bằng một kiến trúc, không phải bốn bản vá rời:**

1. **Distance Tier degressive** (Phần 4) thay /km cố định bằng 7 bậc giảm dần — vừa nâng thu nhập tài xế/quãng ngắn-trung bình, vừa tự động bẻ cong đường cong giá ở quãng dài (giải quyết luôn vấn đề trần giá mà không cần "cắt cứng" giá).
2. **Time Pricing tách 3 lớp** (Moving/Traffic/Waiting — Phần 5) làm rõ tài xế được trả đúng cho cái gì, không gộp mập mờ như hiện tại.
3. **Toàn bộ hệ thống config-driven** (Phần 19) — không một con số nào hardcode trong logic, đúng nguyên tắc Rule Engine đã có ở `ECONOMY_ENGINE.md` Part 11 nhưng lần này áp dụng triệt để cho chính Pricing Service.
4. **Market Index** (Phần 14) làm lớp calibration định kỳ, tách biệt khỏi công thức cước lõi — điều chỉnh vị thế cạnh tranh không cần sửa code.

**Con số minh hoạ xuyên suốt tài liệu** (Car/Standard, ví dụ đại diện): ở kiến trúc V3, một chuyến 60km giờ có giá **473.080đ** — dưới trần 500.000đ (BRB §2.13.6) mà không cần ngoại lệ, so với 603.280đ (vượt trần) ở bảng giá phẳng đề xuất tạm thời tại `MARKET_PRICING_RESEARCH.md` Phần 8.4. Lợi nhuận ròng tài xế Car tăng gần tuyến tính theo quãng đường (31.992đ ở 5km → 164.672đ ở 80km — Phần 16), thay vì gần như dẫm chân tại chỗ như ở BRB hiện hành.

**Điểm giao nhau với thị trường (Phần 15):** với Grab (rẻ nhất trong 3 nền tảng nghiên cứu), đường cong Car V3 giao nhau ở **~11-12km** — dưới ngưỡng này V3 đắt hơn Grab một chút (do bậc /km đầu tiên phải đủ cao để không lặp lại lỗi "quá tuyến tính, không đủ bù chi phí ngắn"), trên ngưỡng này V3 rẻ hơn và khoảng cách nới rộng dần. Đây là đánh đổi có chủ đích, không phải sai sót — trình bày minh bạch, không giấu.

---

## PHẦN 1 — DESIGN PRINCIPLE

### 1.1 Không phải "rẻ nhất" — mà "hợp lý nhất"

"Rẻ nhất" là một cuộc đua một chiều: chỉ có Khách thắng, Driver và Platform đều thua dần (đã phân tích kỹ ở `MARKET_PRICING_RESEARCH.md` Phần 5 — driver Car gần như không có lợi nhuận theo km ở giá hiện tại). "Hợp lý nhất" (Fair-Value Pricing) là trạng thái cả ba bên cùng thắng, đo được bằng số, không phải khẩu hiệu:

| Bên | "Thắng" nghĩa là gì, đo bằng gì |
|---|---|
| **Khách (Customer)** | Trả giá **thấp hơn thị trường một khoảng nhất quán, giải thích được** (8-15% tuỳ hạng xe, không phải ngẫu nhiên) — không phải giá rẻ bất thường không giải thích được lý do, cũng không phải giá ngang/đắt hơn đối thủ. Đo bằng **Customer Saving %** (Phần 2). |
| **Driver** | Lợi nhuận ròng (sau xăng, khấu hao) **tăng theo quãng đường**, không phải dẫm chân tại chỗ — và thu nhập cạnh tranh được với việc chạy cho Grab/Be cùng thời điểm. Đo bằng **Driver Income Index** (Phần 2). |
| **Platform** | Biên đóng góp dương ổn định (không phải biên "vừa đủ sống" mỏng đến mức một cú sốc chi phí nhỏ đẩy về âm — đã xảy ra ở BRB hiện hành, xem `MARKET_PRICING_RESEARCH.md` Phần 10). Đo bằng **Contribution Margin** (Phần 2, 17). |

### 1.2 Ba ràng buộc bất biến kế thừa từ BRB/PRICING_STRATEGY (không đổi ở V3)

1. Surge không vượt ×2.0 (BRB §2.13.3).
2. Giá hiển thị trước khi xác nhận = giá khách trả (BRB §1.2 Nguyên tắc 1).
3. Tài xế luôn tính hoa hồng trên giá đầy đủ trước khuyến mãi (BRB §6.5).

V3 **thêm** một ràng buộc thứ tư, rút ra trực tiếp từ phát hiện của `MARKET_PRICING_RESEARCH.md`:

4. **Không một mức /km cố định duy nhất nào được dùng cho toàn bộ quãng đường của một hạng xe** — đây là nguyên nhân kỹ thuật gốc của cả vấn đề thu nhập tài xế quãng dài lẫn vấn đề vi phạm trần giá. Distance Tier (Phần 4) là ràng buộc kiến trúc, không phải tuỳ chọn.

---

## PHẦN 2 — OBJECTIVE (KPI)

| KPI | Định nghĩa | Mục tiêu V3 | Nguồn đối chiếu |
|---|---|---|---|
| **Driver Income Index** | Thu nhập ròng tài xế/giờ hoạt động, so với trung bình Grab+Be+GreenSM cùng điều kiện | ≥ 100% (bằng hoặc cao hơn thị trường) — đảo ngược phát hiện `MARKET_PRICING_RESEARCH.md` (chỉ 50.7% "tài xế thắng đối thủ" ở cấu hình cũ theo `PRICING_SIMULATION_REPORT.md`) | `PRICING_SIMULATION_REPORT.md` §4.2 |
| **Acceptance Rate** | % lời mời chuyến được tài xế nhận | ≥ 85% (BRB §9.1 không định lượng mục tiêu cụ thể — đây là KPI mới V3 đề xuất) | BRB §9.1 (định nghĩa), mục tiêu V3 mới |
| **Platform Margin** | Contribution Margin / Customer Total, bình quân | 12-18% (dải đã quan sát ở Phần 6/17, đủ đệm cho sốc chi phí, không tối đa hoá — đúng PS §1 "không tối đa lợi nhuận") | `MARKET_PRICING_RESEARCH.md` Phần 6 |
| **Customer Saving** | % rẻ hơn Market Reference (TB Grab/Be/GreenSM) | Bike/XL 8-12%, Car 10-15% (giữ nguyên mục tiêu đã duyệt hướng ở `MARKET_PRICING_RESEARCH.md` Phần 7) | `MARKET_PRICING_RESEARCH.md` Phần 4/7 |
| **Repeat Rate** | % rider có ≥2 chuyến trong 30 ngày | ≥ 40% sau 90 ngày launch (Phần 23) | Mới — chưa có baseline đo thật, cần Analytics service |
| **CAC** (Customer Acquisition Cost) | Chi phí Promotion dành cho First Ride/Referral / số rider mới có ≥1 chuyến | Theo dõi, không đặt trần cứng ở V3.1 — so sánh với LTV (BRB §3.1 "LTV biện minh cho CPA cao") | ECONOMY_ENGINE §10.2 |
| **LTV** (Lifetime Value) | Net Revenue kỳ vọng trọn vòng đời rider | CAC phải < LTV đáng kể (nguyên tắc, chưa có số mục tiêu — cần dữ liệu thật sau launch) | ECONOMY_ENGINE §10.2 |
| **Contribution Margin** | Customer Total − chi phí biến đổi trực tiếp (Phần 17) | Dương ở ≥ 95% chuyến (ngoại trừ loss-leader có chủ đích: voucher sâu, bike ≤2km, Long Pickup xa) | ECONOMY_ENGINE §10.2, `PRICING_SIMULATION_REPORT.md` §Phần 7 |

**Nguyên tắc sắp xếp ưu tiên khi các KPI xung đột:** giữ đúng thứ tự PRICING_STRATEGY §Tóm tắt (Rider → Driver → Tần suất → Cạnh tranh → Hoà vốn → Lợi nhuận) — Driver Income Index không bao giờ bị hy sinh để tối ưu Platform Margin, đúng như PS đã cam kết.

---

## PHẦN 3 — PRICING COMPONENT (thiết kế lại toàn bộ)

| Thành phần | Vai trò trong V3 | Thay đổi so với BRB hiện hành |
|---|---|---|
| **Base Fare** | Phí cố định mở chuyến, theo hạng xe + thành phố (Phần 6), không nhân surge | Giữ nguyên vai trò, chỉ đổi giá trị (Phần 4) và thêm hệ số thành phố |
| **Distance Fare** | Tổng của 7 bậc quãng đường (Phần 4), mỗi bậc một đơn giá riêng | **Thay đổi cấu trúc** — không còn 1 mức /km, đây là thay đổi lớn nhất của V3 |
| **Time Fare** | Tách 3 lớp: Moving (không tính riêng) / Traffic (tốc độ <10km/h khi đang chở khách) / Waiting (trước khi chuyến bắt đầu) — Phần 5 | Làm rõ ranh giới 3 khái niệm đang gộp mờ trong BRB §2.2.3/§2.2.9 |
| **Booking Fee** | Phí dịch vụ nền tảng cố định, không surge, 100% platform | Tăng nhẹ theo tỷ trọng (đã đề xuất ở `MARKET_PRICING_RESEARCH.md` §8.3), không đổi vai trò |
| **Airport Fee** | Tách 4 thành phần: Queue/Pickup/Dropoff/Priority (Phần 7) | **Thay đổi cấu trúc** — không còn 1 phí flat duy nhất |
| **Bridge Fee** | Pass-through 100%, 0% hoa hồng, tài xế khai báo + xác thực | Giữ nguyên (đã là **[MỚI — cần tu chính BRB]** từ PRICING_STRATEGY §2.2, V3 kế thừa) |
| **Tunnel Fee** — **[MỚI]** | Pass-through 100%, 0% hoa hồng, cùng cơ chế Bridge/Toll — hầm (ví dụ hầm Thủ Thiêm, hầm Hải Vân) có phí riêng biệt về vận hành (giới hạn tốc độ, kiểm soát ra vào) nên tách khỏi Bridge dù cùng bản chất kinh tế, để báo cáo minh bạch theo loại hạ tầng | Chưa tồn tại ở BRB/PRICING_STRATEGY — đề xuất mới của V3 |
| **Waiting Fee** | Sau 3 phút miễn phí kể từ "Đã đến", theo /phút từng hạng xe | Giữ nguyên cơ chế, xem lại mức /phút theo Phần 5 |
| **Long Pickup Compensation** | Nền tảng chịu 100%, rider không trả thêm | Giữ nguyên nguyên tắc, không đổi (đã đúng thiết kế PRICING_STRATEGY §2.2) |
| **Cancellation Fee** | 80/20 tài xế/nền tảng, theo thời điểm huỷ | Giữ nguyên (BRB §10.1) |
| **No Show Fee** — **[MỚI]** | Rider không xuất hiện sau khi tài xế chờ vượt ngưỡng (Waiting grace + X phút, X theo hạng xe) → tài xế được huỷ chuyến, thu No Show Fee (cao hơn Cancellation Fee một bậc vì tài xế đã chờ lâu hơn, tổn thất thời gian lớn hơn) | Chưa tồn tại — hiện BRB chỉ có Cancellation Fee, không phân biệt "rider huỷ chủ động" với "rider biến mất không huỷ" |
| **Dynamic Pricing (Surge)** | Rule Engine tất định theo DSR, trần ×2.0 (Phần 8) | Giữ nguyên cơ chế lõi, mở rộng danh mục tín hiệu đầu vào |
| **Promotion** | Một chồng (stack) có luật rõ ràng cộng dồn/loại trừ (Phần 9) | Formalize luật đã tồn tại trong `promotion_service.go` thành tài liệu kiến trúc |
| **Membership** | Không đổi công thức giá, chỉ đổi dịch vụ (Phần 10) | Giữ nguyên nguyên tắc bất biến đã có ở ECONOMY_ENGINE §8.1 |
| **Coupon** | Một dạng con của Promotion (mã nhập tay, `Voucher.Code` khác rỗng) | Không phải thành phần giá riêng — đã là 1 phần của Promotion Stack |
| **Platform Subsidy** | Phần nền tảng tự chi trả để bù Minimum Driver Earning Guarantee, Long Pickup, top-up khi Voucher vượt giá trị chuyến | Formalize thành 1 dòng riêng trong Revenue Distribution (Phần 11), hiện đang ẩn trong "Net Driver" |
| **Driver Bonus** | Quest/Streak/Peak/Airport/Rain/Referral (BRB Part 8) | Không phải thành phần giá — chi từ ngân sách Incentive riêng, không hiển thị trong giá rider nhìn thấy (giữ nguyên) |

---

## PHẦN 4 — DISTANCE TIER

### 4.1 Nguyên tắc

Thay `PerKmRate` (1 số duy nhất) bằng `DistanceTiers` (mảng 7 bậc, mỗi bậc có khoảng km và đơn giá riêng, **giảm dần**). Đây là thiết kế **cước bậc thang giảm dần** (degressive tariff) — đúng pattern cả 5 đối thủ nghiên cứu ở `MARKET_PRICING_RESEARCH.md` Phần 2 đều dùng (Grab/Be/GreenSM/Mai Linh/Vinasun tất cả đều rẻ hơn ở km xa so với km đầu).

**Lý do kỹ thuật:** bậc đầu (0-2km) phải đủ cao để bù chi phí khởi hành + xăng/khấu hao không được Base Fare bù đủ ở chuyến rất ngắn (đây là gốc rễ vấn đề "chuyến ngắn lỗ" mà `PRICING_SIMULATION_REPORT.md` đã phát hiện cho xe máy). Bậc cuối (60km+) phải đủ thấp để đường cong không vượt trần giá tuyệt đối và cạnh tranh được ở các chuyến liên tỉnh/sân bay xa.

### 4.2 Bảng bậc đề xuất (VND/km, Car/Standard)

| Bậc | Khoảng | Đơn giá /km | So với /km phẳng cũ (8.600đ) |
|---|---|---|---|
| 1 | 0-2km | 9.500 | +10.5% |
| 2 | 2-5km | 8.600 | 0% (mốc neo) |
| 3 | 5-10km | 7.800 | -9.3% |
| 4 | 10-20km | 7.000 | -18.6% |
| 5 | 20-40km | 6.200 | -27.9% |
| 6 | 40-60km | 5.400 | -37.2% |
| 7 | 60km+ | 4.600 | -46.5% |

Bike và XL dùng cùng shape (7 bậc, tỷ lệ giảm dần tương tự), scale theo tỷ lệ đã thiết lập ở `MARKET_PRICING_RESEARCH.md` (Bike ≈ 42% Car, XL ≈ 134% Car):

| Bậc | Khoảng | Bike /km | XL /km |
|---|---|---|---|
| 1 | 0-2km | 4.200 | 12.500 |
| 2 | 2-5km | 3.700 | 11.300 |
| 3 | 5-10km | 3.300 | 10.200 |
| 4 | 10-20km | 3.000 | 9.100 |
| 5 | 20-40km | 2.700 | 8.100 |
| 6 | 40-60km | 2.400 | 7.000 |
| 7 | 60km+ | 2.100 | 6.000 |

### 4.3 Ví dụ tính (Car, minh hoạ cách bậc cộng dồn)

Chuyến 25km: 2km×9.500 + 3km×8.600 + 5km×7.800 + 15km×7.000 = 19.000 + 25.800 + 39.000 + 105.000 = **188.800đ** Distance Fare (so với 25km×8.600 = 215.000đ nếu dùng /km phẳng cũ — rẻ hơn 12% ở mốc 25km, đúng hướng "quãng dài rẻ hơn tương đối" mà không cắt giảm quãng ngắn).

### 4.4 Base Fare + Minimum Fare mới (không đổi vai trò, chỉ đối chiếu lại với bậc 1)

| | Car | XL | Bike |
|---|---|---|---|
| Base Fare | 13.000 | 22.000 | 2.500 |
| Minimum Fare | 30.000 | 48.000 | 9.000 |
| Booking Fee | 3.000 | 3.000 | 1.000 |

(Giữ nguyên giá trị đã thiết kế ở `MARKET_PRICING_RESEARCH.md` §8.2 — chỉ thay cấu trúc Distance Fare, không thay đổi 3 tham số này vì chúng không phải nguyên nhân của vấn đề trần giá, xem `MARKET_PRICING_RESEARCH.md` Phần 1.1.)

---

## PHẦN 5 — TIME PRICING

### 5.1 Ba lớp thời gian, ba mục đích khác nhau

| Lớp | Khi nào tính | Đơn giá (Car) | Vì sao tách riêng |
|---|---|---|---|
| **Moving Time** | Đang chở khách, tốc độ ≥ 10km/h | **Không tính riêng** — đã được Distance Fare bù (đúng nguyên tắc BRB §2.2.3 "loại trừ lẫn nhau trong cùng một giây") | Tránh tính hai lần cùng một hiện tượng di chuyển — giữ nguyên logic đã đúng của BRB, chỉ đặt tên rõ ràng hơn ("Moving Time" thay vì để ẩn trong công thức) |
| **Traffic Time** (đổi tên từ "Time Fare" cho rõ nghĩa) | Đang chở khách, tốc độ < 10km/h (kẹt xe) | 540đ/phút (Car) — **đây chính là "Time Fare" hiện tại của BRB §2.2.3, chỉ đổi tên cho khớp đúng bản chất** | Tên "Time Fare" hiện tại dễ nhầm là tính cho MỌI phút của chuyến — thực ra chỉ tính khi kẹt xe. Đổi tên tránh rider/tài xế hiểu sai khi đọc bảng minh bạch giá (BRB §1.2 "Rules Are Public" đòi hỏi tên gọi rõ nghĩa) |
| **Waiting Time** | Trước khi chuyến bắt đầu, sau khi tài xế bấm "Đã đến", vượt quá 3 phút miễn phí | 500đ/phút (Car, giữ nguyên BRB §2.2.9) | Khác bản chất với Traffic Time: đây là thời gian tài xế **chưa** chở khách, chờ rider chuẩn bị — chi phí cơ hội khác nhau, cần tách để báo cáo minh bạch đúng "cái gì được trả cho cái gì" |

### 5.2 No Show — ranh giới với Waiting/Cancellation (Phần 3)

Nếu Waiting Time vượt một ngưỡng thứ hai (đề xuất: 3 phút miễn phí + 7 phút tính phí + quá phút thứ 10 tổng cộng → tài xế được chủ động huỷ, thu No Show Fee) — ranh giới giữa "Waiting Fee tiếp tục cộng dồn" và "chuyển sang No Show" cần một mốc thời gian rõ ràng để tránh tài xế chờ vô thời hạn.

---

## PHẦN 6 — CITY COEFFICIENT

### 6.1 Thiết kế

Một hệ số nhân (`CityCoefficient`) áp lên **toàn bộ** Base + Distance + Time (không áp lên Booking Fee, Bridge/Tunnel pass-through — giữ nguyên nguyên tắc "phí cố định không surge/không hệ số vùng miền" đã có cho Booking Fee), để một bảng `DistanceTiers` gốc (Phần 4, hiệu chỉnh theo TP.HCM) có thể tái sử dụng cho các thành phố khác mà không cần một bảng giá hoàn toàn riêng biệt.

| Thành phố | CityCoefficient | Căn cứ (định tính — cần dữ liệu thị trường riêng từng thành phố trước khi chốt số, đây là khởi điểm ước lượng) |
|---|---|---|
| **TP.HCM** | 1.00 (mốc neo — nơi nghiên cứu thị trường Phần 2 `MARKET_PRICING_RESEARCH.md` được thực hiện) | Thị trường lớn nhất, cạnh tranh khốc liệt nhất, chi phí sinh hoạt/xăng cao nhất |
| **Hà Nội** | 1.00 | Quy mô thị trường tương đương TP.HCM, cùng nhóm cạnh tranh (Grab/Be/Xanh SM đều có mặt đầy đủ) |
| **Đà Nẵng** | 0.90 | Thị trường trung bình, chi phí sinh hoạt thấp hơn ~10% |
| **Hải Phòng** | 0.85 | Thị trường nhỏ hơn Đà Nẵng, cạnh tranh ít gay gắt hơn |
| **Cần Thơ** | 0.82 | Thị trường nhỏ nhất trong nhóm, chi phí vận hành thấp nhất |

### 6.2 Khả năng mở rộng

`CityCoefficient` là một dòng trong bảng config (Phần 19), không phải hằng số trong code — thêm thành phố mới chỉ cần thêm một dòng, không cần release ứng dụng (đúng nguyên tắc Rule Engine "Scoped theo thành phố" đã có ở ECONOMY_ENGINE §11.2). Mọi hệ số ở bảng trên là **khởi điểm ước lượng**, cần Pricing/Finance khảo sát riêng từng thị trường trước khi ra mắt chính thức tại thành phố đó (đúng cách `MARKET_PRICING_RESEARCH.md` đã khảo sát riêng cho TP.HCM).

---

## PHẦN 7 — AIRPORT (thiết kế lại toàn diện, không chỉ 1 phí)

`MARKET_PRICING_RESEARCH.md` Phần 1.3 đã phát hiện: Airport Fee flat 10.000đ áp đồng nhất mọi hạng xe tạo bất cân xứng (đẩy giá bike sân bay lên +56% so với thị trường vì đối thủ không thu phụ phí sân bay cho xe máy). V3 tách thành 4 thành phần độc lập:

| Thành phần | Áp dụng | Ai trả | Vai trò |
|---|---|---|---|
| **Airport Queue Compensation** | Chỉ Car/XL — tài xế xếp hàng tại bãi chờ sân bay trước khi được điều phối đón khách | **Nền tảng chi trả**, không tính vào giá rider (giống Long Pickup Compensation — chi phí do đặc thù vận hành sân bay, không phải lỗi/lựa chọn của rider) | Bù thời gian chờ thực tế của tài xế tại bãi, tách khỏi Waiting Time (vốn tính từ lúc "Đã đến" điểm đón cụ thể, không phải từ lúc vào bãi chờ chung) |
| **Airport Pickup Fee** | Mọi hạng xe **trừ Bike** (đúng thực tế thị trường quan sát được — Grab/Be/GreenSM không phụ phí sân bay cho xe máy) | Rider | Thay thế Airport Fee flat cũ; mức đề xuất: Car 15.000đ, XL 20.000đ, Bike **0đ** |
| **Airport Dropoff Fee** | Mọi hạng xe trừ Bike, mức thấp hơn Pickup (không cần xếp hàng khi trả khách) | Rider | Đề xuất: Car 5.000đ, XL 7.000đ, Bike 0đ |
| **Airport Priority Dispatch** | Driver Tier Gold+ (BRB §7.5 Priority Dispatch, áp dụng cho khu vực sân bay) | Không phải phí — quyền lợi dịch vụ | Tài xế tier cao được ưu tiên nhận chuyến sân bay trước — không ảnh hưởng giá, chỉ ảnh hưởng thứ tự điều phối |

**Kết quả:** một chuyến Bike sân bay không còn phụ phí nào (khớp đúng thị trường), một chuyến Car sân bay chịu 15.000đ (pickup) hoặc 5.000đ (dropoff) thay vì 10.000đ đồng nhất — tổng thể gần với chi phí thực tế sân bay hơn (đón thường mất thời gian xếp hàng hơn trả).

---

## PHẦN 8 — SURGE (Rule Engine, không AI)

### 8.1 Giữ nguyên lõi tất định, mở rộng danh mục tín hiệu

Lõi DSR-band (BRB §2.13.2, trần ×2.0 BRB §2.13.3) **không đổi** — đây là cơ chế đã đúng. V3 chuẩn hoá **cách mỗi tín hiệu tương tác với DSR** thành một bảng tường minh, để không cần đoán "cái nào nhân, cái nào cộng, cái nào loại trừ nhau":

| Tín hiệu | Loại | Tương tác với Surge (DSR-based) |
|---|---|---|
| **Demand** | Đầu vào tính DSR | DSR = requests/drivers hoạt động trong zone — không phải phụ phí riêng |
| **Supply** | Đầu vào tính DSR | Cùng công thức DSR ở trên |
| **Weather (Rain)** | Hệ số nhân cố định (×1.15) | Cộng dồn tuần tự VỚI Surge (không thay thế) theo đúng thứ tự BRB §2.17: Surge trước, Night/Holiday/Rain sau, trần cộng dồn ×1.60 cho nhóm tĩnh |
| **Holiday** | Hệ số nhân cố định (×1.15) | Cùng nhóm với Weather, trần cộng dồn ×1.60 |
| **Night** | Hệ số nhân cố định (×1.20) | Cùng nhóm |
| **Peak Hour** | Hệ số nhân cố định (×1.10) | **Loại trừ lẫn nhau với Surge** (BRB §2.2.12 — surge thắng khi cả hai active, không cộng dồn) |
| **Airport** | Phụ phí cố định (Phần 7), không phải hệ số | Cộng dồn **sau** khi Surge + nhóm tĩnh đã áp lên (base+distance+time), không tự nhân với surge (giữ nguyên PRICING_STRATEGY §5.3) |
| **Event/Festival** | Geofence tạm thời + phụ phí cố định giống Airport (PS §5.2 "Event Zone Surcharge", đã là **[MỚI — cần tu chính BRB]**) | Cùng nhóm với Airport — cộng dồn cố định, không nhân surge |
| **Emergency** — **[MỚI]** | Cờ vận hành thủ công (Operations bật/tắt) | **Override an toàn**: khi Emergency bật (thiên tai, ngập lụt diện rộng, sự cố an ninh khu vực), Surge bị **khoá cứng về ×1.0** bất kể DSR thực tế — ưu tiên Fairness First (Constitution §1.2) hơn cân bằng cung-cầu trong tình huống khẩn cấp, tránh lặp lại khủng hoảng truyền thông "surge tăng vọt lúc khẩn cấp" mà Uber từng gặp (PRICING_STRATEGY §0.2) |

### 8.2 Vì sao vẫn Rule Engine, không AI

Đúng nguyên tắc PRICING_STRATEGY §5.1: mọi tín hiệu ánh xạ sang bảng tra cứu công khai — một rider có thể tự tính ra giá của họ. Thêm tín hiệu Emergency **củng cố** lý do này: một override an toàn dạng "nếu X thì khoá về 1.0" chỉ có thể verify được (bởi CPO, bởi báo chí, bởi cơ quan quản lý) nếu nó là luật tường minh — một mô hình AI "quyết định không surge" không thể chứng minh được là nó LUÔN LUÔN làm đúng trong tình huống khẩn cấp.

---

## PHẦN 9 — PROMOTION STACK

### 9.1 Mô hình hiện tại (đã có thật trong `promotion_service.go`, không phải thiết kế mới)

`PromotionService.Evaluate()` áp dụng đúng 1 quy tắc: trong số mọi voucher đủ điều kiện, chọn **đúng 1** theo thứ tự Priority → Discount cao hơn → Hết hạn sớm hơn (BRB §3.4/§3.5), tối đa 1 voucher/chuyến (BRB §4.7). `Voucher.Stackable` tồn tại như một cờ nhưng **chưa được triển khai** (TODO tường minh trong code) — đây là điểm V3 cần quyết định rõ.

### 9.2 Ma trận cộng dồn/loại trừ đề xuất cho V3

| Cặp | Cộng dồn hay loại trừ? | Lý do |
|---|---|---|
| Voucher ↔ Voucher khác (2 mã cùng lúc) | **Loại trừ** (giữ nguyên BRB §4.7) | Tránh chồng chiết khấu không kiểm soát ngân sách — đây là ràng buộc nền tảng, không nên mở trừ khi có phê duyệt CPO từng cặp cụ thể (`Stackable` field đã chừa chỗ cho ngoại lệ CPO duyệt, không phải mặc định mở) |
| Voucher (upfront) ↔ Cashback | **Cộng dồn** (đã đúng, BRB §3.4 #7 — Cashback là hậu-chuyến, độc lập hoàn toàn) | Cashback không đi qua `Evaluate()`, không cạnh tranh ngân sách/độ ưu tiên với voucher trước-chuyến |
| Voucher ↔ Membership | **Cộng dồn** | Membership không phải một discount type trong `PromotionType` theo nghĩa tạo giảm giá cước — nó chỉ giảm Booking Fee/ưu tiên dịch vụ (Phần 10), nên không "cạnh tranh vị trí" với Voucher trong `Evaluate()` |
| Voucher ↔ Wallet (phương thức thanh toán) | **Cộng dồn** | Wallet là phương thức thanh toán, không phải khuyến mãi — không đi qua Promotion Engine |
| Referral ↔ First Ride | **Loại trừ** (cùng nhóm ưu tiên trong bảng BRB §3.4, giá trị cao hơn thắng) | Cả hai đều nhắm rider mới — không có lý do kinh doanh để cộng dồn hai ưu đãi "chuyến đầu tiên" |
| Campaign (Coupon) ↔ Golden Hour/Rain/Weekend | **Loại trừ**, trừ khi CPO duyệt riêng cặp cụ thể qua `Stackable` | Giữ nguyên tắc ngân sách kiểm soát được — mở rộng `Stackable` thành cơ chế **cấu hình theo cặp** (không phải cờ nhị phân đơn) là việc cần làm ở V3.2 (Phần 20) |

### 9.3 Việc cần làm để hiện thực hoá `Stackable` (không code ngay, chỉ ghi nhận thiết kế)

`Stackable` hiện là 1 boolean trên từng Voucher — không đủ biểu diễn "voucher A được stack với voucher B nhưng không với C". V3.2 cần một bảng cấu hình `stackable_pairs (voucher_type_a, voucher_type_b, approved_by, approved_at)` thay vì 1 cờ đơn — đây là thay đổi schema, liệt kê ở Phần 21 Migration Plan.

---

## PHẦN 10 — MEMBERSHIP

Kế thừa nguyên vẹn thiết kế đã có ở `ECONOMY_ENGINE.md` Phần 8 (Free/Silver/Gold/Diamond, điều kiện dựa trên hành vi) — nguyên tắc bất biến nhắc lại: **Membership không bao giờ đổi công thức giá cước** (Base/Distance/Time/Minimum không đổi theo hạng thành viên), chỉ đổi:

| Hạng | Quyền lợi (dịch vụ, không phải giá) |
|---|---|
| Free | Không có |
| Silver | Ưu tiên hỗ trợ nhanh hơn |
| Gold | Ưu tiên ghép chuyến nhẹ (Priority Dispatch biên độ nhỏ) |
| Diamond | Toàn bộ Gold + **trần surge cá nhân thấp hơn trần chung** (ví dụ ×1.5 thay vì ×2.0 — đây là giới hạn giá CAO NHẤT phải trả khi có surge, không phải giá THẤP HƠN ở điều kiện thường, nên không vi phạm nguyên tắc "cùng giá cho chuyến không surge") |

**Lưu ý đã có ở ECONOMY_ENGINE §8.2, nhắc lại vì quan trọng:** ngoại lệ trần surge cho Diamond cần **Legal + CPO** xác nhận rõ ràng không vi phạm tinh thần Rider Fairness trước khi triển khai — V3 không tự động coi đây là đã duyệt.

---

## PHẦN 11 — REVENUE DISTRIBUTION (mỗi cuốc)

Waterfall đầy đủ, dùng số ví dụ Car 20km (đã tính ở `MARKET_PRICING_RESEARCH.md` Phần 5/6, tái sử dụng để nhất quán số liệu giữa 2 tài liệu):

```
Khách trả (Customer Total)                     211.760đ
  │
  ├─ (nếu có Voucher — Phần 9)                  ví dụ -20% = -41.752đ (Platform Subsidy, không trừ vào phần Driver)
  │
  ▼
Ride Fare (Base+Distance+Time+phụ phí, sau khi trừ Voucher, trước Booking Fee)
  │
  ├──▶ Commission (20% Bronze → 12% Diamond, BRB §7.1)      41.752đ (Bronze, trên giá TRƯỚC voucher — BRB §6.5)
  ├──▶ Net Driver (= Ride Fare − Commission, sàn 20.000đ)    167.008đ
  │        ├─ Xăng                                          40.000đ
  │        ├─ Khấu hao                                      36.000đ
  │        └─ Lợi nhuận ròng tài xế                          91.008đ
  │
  ├──▶ Booking Fee (100% platform)                            3.000đ
  ├──▶ VAT (10%, trên Commission+Booking Fee, ASSUMPTION)     4.475đ
  ├──▶ Phí Gateway (1.8% Customer Total, ASSUMPTION)          3.812đ
  ├──▶ Opex cố định/cuốc (SMS/Map/Cloud/Support, ASSUMPTION)    880đ
  │
  ▼
Net Margin (Platform Profit)                                35.585đ  (biên 16.8% trên Customer Total)
```

**Khác biệt so với waterfall trong `MARKET_PRICING_RESEARCH.md`:** tài liệu đó gộp "Platform Subsidy" ẩn bên trong so sánh sensitivity (Phần 10 của tài liệu đó). V3 **đặt Platform Subsidy thành một dòng tường minh** trong revenue distribution — vì đây là tiền thật chảy ra khỏi Promotion Fund (ECONOMY_ENGINE §3.5), cần nhìn thấy ngay trong luồng tiền chính, không chỉ trong phân tích sensitivity riêng.

---

## PHẦN 12 — LONG DISTANCE (Distance Discount Curve)

Đây chính là Phần 4 nhìn từ góc độ vấn đề cần giải quyết: **Distance Tier degressive chính là Distance Discount Curve.** Không cần một cơ chế "giảm giá quãng dài" tách biệt (như một loại Promotion) — độ dốc giảm dần đã nằm ngay trong cấu trúc `DistanceTiers` (Phần 4.2), không phải một lớp chiết khấu áp thêm vào sau.

**Vì sao thiết kế thành Tier thay vì một công thức giảm liên tục (ví dụ hàm mũ giảm dần theo km)?** Ba lý do:
1. **Giải thích được trong 60 giây** (BRB §2.1) — "km 40-60 tính 5.400đ/km" dễ hiểu hơn "per_km × e^(-0.003×km)".
2. Khớp đúng cách **cả 5 đối thủ nghiên cứu đều làm** (bậc rời rạc, không phải hàm liên tục) — không tạo trải nghiệm khác lạ.
3. Rule Engine (Phần 19) cấu hình bậc rời rạc dễ audit/thay đổi độc lập từng bậc hơn là chỉnh tham số một hàm toán học trừu tượng.

So sánh trực quan (Car, đã tính ở Phần 4.3 và script `v3_tiers.py`):

| Km | Distance Fare (Tier V3) | Distance Fare (nếu /km phẳng 8.600đ cũ) | Tiết kiệm |
|---|---|---|---|
| 10 | 83.800 | 86.000 | -2.6% |
| 25 | 188.800 | 215.000 | -12.2% |
| 40 | 277.800 | 344.000 | -19.2% |
| 60 | 385.800 | 516.000 | -25.2% |
| 100 | 569.800 | 860.000 | -33.7% |

→ Đường cong tự động "bẻ cong" mà không cần bất kỳ luật đặc biệt "nếu km > 50 thì..." nào — đây là lý do tại sao Phần 8.4 của `MARKET_PRICING_RESEARCH.md` (vi phạm trần giá) tự động biến mất ở V3 (xem xác nhận số ở Tóm tắt điều hành: 60km = 473.080đ, dưới trần).

---

## PHẦN 13 — PRICE CEILING

| Ràng buộc | Mức đề xuất | Áp dụng cho | Căn cứ |
|---|---|---|---|
| **Minimum Fare** | Car 30.000 / XL 48.000 / Bike 9.000 | Ride Fare trước Booking Fee | Đã thiết kế ở `MARKET_PRICING_RESEARCH.md` §8.2, giữ nguyên |
| **Maximum Fare (chuyến nội thành/liên tỉnh ngắn)** | 500.000đ (Car/Standard) — giữ nguyên BRB §2.13.6 | Toàn bộ Ride Fare (base+distance+time+phụ phí+surge), trước Booking Fee | BRB §2.13.6 — **với Distance Tier degressive (Phần 4/12), trần này chỉ còn bị chạm ở quãng ≥ ~75-80km** (tính từ bảng Phần 12: 60km=473.080 chưa chạm, cần nội suy thêm để xác định điểm chạm chính xác — việc này thuộc Phần 18 mô phỏng 1000 kịch bản, không tính tay ở đây) |
| **Maximum Fare (chuyến dài/sân bay liên tỉnh)** — **[MỚI]** | 1.000.000đ | Áp dụng khi quãng đường > 60km HOẶC một đầu là sân bay ngoài thành phố (theo đúng ghi chú đã có sẵn trong BRB §2.13.6 "Airport long-distance trips have a higher cap negotiated separately" — V3 chính thức hoá con số cho ghi chú này) | BRB §2.13.6 (ghi chú), V3 định lượng cụ thể |
| **Maximum Surge** | ×2.0, không ngoại lệ | Toàn bộ hạng xe | BRB §2.13.3, bất biến |
| **Maximum Waiting Fee** | Cap ở 60 phút tính phí (sau đó chuyển No Show — Phần 3/5) | Mọi hạng xe | Mới — tránh Waiting Fee cộng dồn vô hạn nếu rider không phản hồi cũng không để tài xế huỷ |
| **Maximum Airport Queue Compensation** | Cap ở 45 phút chờ (nền tảng chi trả — cần trần để kiểm soát ngân sách, giống nguyên tắc Long Pickup) | Car/XL tại khu vực sân bay | Mới, theo cùng logic kiểm soát ngân sách của Long Pickup Compensation |

---

## PHẦN 14 — MARKET INDEX

### 14.1 Thiết kế (kế thừa khung đã đề xuất ở `MARKET_PRICING_RESEARCH.md` Phần 9, dùng ví dụ minh hoạ của nhiệm vụ lần này làm khung trình bày)

```
Panda_Fare(scenario) = ReferenceMarketFare(scenario) × TargetIndex[hạng xe][thành phố]
```

| Nền tảng | Index minh hoạ (theo ví dụ nhiệm vụ) | Index đo được thật (Car, TB 13 khoảng cách — `MARKET_PRICING_RESEARCH.md` Phần 9.2) |
|---|---|---|
| Grab | 100 (mốc neo) | 100 (mốc neo) |
| Be | 97 | ~123 |
| GreenSM | 95 | ~138 |
| **Panda (mục tiêu)** | **90** | **~87 (đề xuất V3, xem 14.2)** |

**Ghi chú minh bạch:** ví dụ trong nhiệm vụ (Be=97, GreenSM=95 — ngụ ý hai nền tảng này rẻ hơn Grab) **không khớp với số đo thật** đã nghiên cứu (Be/GreenSM thực ra đắt hơn Grab đáng kể ở hạng Car — xem `MARKET_PRICING_RESEARCH.md` Phần 9.2 với nguồn trích dẫn cụ thể). V3 dùng cấu trúc bảng của ví dụ nhưng **giữ số đo thật**, không sửa số đo cho khớp ví dụ minh hoạ — đúng nguyên tắc không bịa số xuyên suốt toàn bộ hệ thống tài liệu Panda.

### 14.2 TargetIndex đề xuất cho Pricing V3 (không đổi so với đề xuất trước, đã kiểm chứng bằng 312 kịch bản)

| Hạng xe | TargetIndex |
|---|---|
| Car | 0.85-0.90 |
| XL | 0.88-0.92 |
| Bike | 0.88-0.92 |

### 14.3 Vai trò trong kiến trúc config (Phần 19)

`ReferenceMarketFare` là một bảng tĩnh (Base/Distance-tier/Time riêng, mô phỏng đường cong thị trường) được Pricing/Finance cập nhật **định kỳ** (đề xuất: hàng quý) qua khảo sát thủ công như đã làm ở `MARKET_PRICING_RESEARCH.md` — **không phải** một lời gọi API sống đến Grab/Be/GreenSM. `TargetIndex` là con số duy nhất Product cần đổi để dịch chuyển toàn bộ vị thế giá — không cần tính lại 7 bậc Distance Tier bằng tay.

---

## PHẦN 15 — REVENUE CURVE (Distance ↔ Fare)

### 15.1 Bảng dữ liệu (Car, điều kiện Nội thành, không surge)

| Km | Grab | Be | GreenSM | **Panda V3** |
|---|---|---|---|---|
| 2 | 29.000 | 43.306 | 35.000 | **37.376** |
| 4 | 49.000 | 66.916 | 65.000 | **56.952** |
| 6 | 69.000 | 90.526 | 95.000 | **75.728** |
| 8 | 89.000 | 114.136 | 125.000 | **93.704** |
| 10 | 109.000 | 137.746 | 155.000 | **111.680** |
| 12 | 129.000 | 159.304 | 185.000 | **128.056** |
| 15 | 159.000 | 191.641 | 230.000 | **152.620** |
| 20 | 209.000 | 245.536 | 305.000 | **193.560** |
| 40 | 409.000 | 461.116 | 557.000 | **341.320** |
| 60 | 609.000 | 676.696 | 797.000 | **473.080** |

### 15.2 Phác hoạ hình dạng đường cong (ASCII, trục Y không theo tỷ lệ tuyệt đối — chỉ minh hoạ thứ tự và độ dốc tương đối)

```
Giá
 │                                                          ╱ GreenSM (dốc nhất)
 │                                                       ╱ ╱
 │                                                    ╱╱╱  Be
 │                                                 ╱╱╱  ╱
 │                                              ╱╱╱   ╱   Grab
 │                                          ___╱╱╱  ╱
 │                                    __--¯¯    ╱ ╱
 │                            __--¯¯¯       ╱ ╱      Panda V3 (thoải nhất
 │                    __--¯¯¯            ╱ ╱          từ ~15km trở đi —
 │            __--¯¯¯                 ╱╱               Distance Tier bậc 5-7)
 │     __--¯¯¯  ← điểm giao (~11-12km)
 └────────────────────────────────────────────────────────────── Km
     2   6   10  12  15    20        40              60
```

### 15.3 Điểm giao nhau (crossover point)

So với **Grab** (rẻ nhất trong 3 nền tảng công nghệ nghiên cứu): Panda V3 đắt hơn Grab ở quãng < ~11-12km (ví dụ 2km: Panda 37.376 vs Grab 29.000, **+28.9%**), và **giao nhau ở khoảng 11-12km** (12km: Panda 128.056 vs Grab 129.000, chênh -0.7%), sau đó rẻ hơn dần: 20km rẻ hơn 7.4%, 40km rẻ hơn 16.6%, 60km rẻ hơn 22.3%.

**Đây là đánh đổi có chủ đích, không phải lỗi thiết kế:** bậc 1 (0-2km) phải đủ cao (9.500đ/km) để tránh lặp lại lỗi "chuyến ngắn lỗ có cấu trúc" mà `PRICING_SIMULATION_REPORT.md` đã phát hiện cho xe máy — cái giá phải trả là Panda không còn rẻ hơn Grab ở **mọi** cự ly, chỉ rẻ hơn từ khoảng giữa trở đi. So với **Be và GreenSM** (đắt hơn Grab đáng kể — Phần 9.2), Panda V3 vẫn rẻ hơn ở **hầu hết** mọi cự ly kể cả ngắn (2km: Panda 37.376 < Be 43.306, xấp xỉ GreenSM 35.000).

**Khuyến nghị theo dõi sau launch (không tự quyết ở đây):** nếu dữ liệu thật cho thấy chuyến < 8km chiếm tỷ trọng lớn và rider nhạy cảm với việc "đắt hơn Grab ở chuyến ngắn", cân nhắc hạ bậc 1 xuống ~8.800-9.000đ/km ở V3.1 (Phần 20), đánh đổi bằng biên lợi nhuận mỏng hơn một chút ở chuyến ngắn — đây là quyết định cần dữ liệu thật, không quyết ở giai đoạn thiết kế.

---

## PHẦN 16 — DRIVER ECONOMICS

Cùng phương pháp/giả định chi phí như `MARKET_PRICING_RESEARCH.md` Phần 5 (xăng 25.000đ/lít; tiêu hao Bike 2.0L/Car 8.0L/XL 9.5L trên 100km; khấu hao Bike 500đ/Car 1.800đ/XL 2.200đ mỗi km — ASSUMPTION, chưa CFO xác nhận), tính lại với Distance Tier V3 (Phần 4), mở rộng thêm mốc 80km theo đúng yêu cầu:

| Xe | Km | Khách trả | Net Driver (Bronze 20%) | Xăng | Khấu hao | **Lợi nhuận ròng tài xế** |
|---|---|---|---|---|---|---|
| Bike | 5 | 25.530 | 20.000 (sàn) | 2.500 | 2.500 | **15.000** |
| Bike | 10 | 44.560 | 34.848 | 5.000 | 5.000 | **24.848** |
| Bike | 20 | 79.620 | 62.896 | 10.000 | 10.000 | **42.896** |
| Bike | 40 | 143.740 | 114.192 | 20.000 | 20.000 | **74.192** |
| Bike | 80 | 253.980 | 202.384 | 40.000 | 40.000 | **122.384** |
| Car | 5 | 66.740 | 50.992 | 10.000 | 9.000 | **31.992** |
| Car | 10 | 111.680 | 86.944 | 20.000 | 18.000 | **48.944** |
| Car | 20 | 193.560 | 152.448 | 40.000 | 36.000 | **76.448** |
| Car | 40 | 341.320 | 270.656 | 80.000 | 72.000 | **118.656** |
| Car | 80 | 588.840 | 468.672 | 160.000 | 144.000 | **164.672** |
| XL | 5 | 91.600 | 70.880 | 11.875 | 11.000 | **48.005** |
| XL | 10 | 150.300 | 117.840 | 23.750 | 22.000 | **72.090** |
| XL | 20 | 256.700 | 202.960 | 47.500 | 44.000 | **111.460** |
| XL | 40 | 449.500 | 357.200 | 95.000 | 88.000 | **174.200** |
| XL | 80 | 771.100 | 614.480 | 190.000 | 176.000 | **248.480** |

**So với BRB hiện hành (`MARKET_PRICING_RESEARCH.md` Phần 5.1):** lợi nhuận ròng tài xế Car ở 40km tăng từ 12.160đ (BRB hiện hành) lên **118.656đ** (V3) — gấp ~9.8 lần, phản ánh đúng việc /km đã được hiệu chỉnh sát chi phí thực + tăng đúng theo Distance Tier thay vì gần như dẫm chân tại chỗ.

---

## PHẦN 17 — PLATFORM ECONOMICS

### 17.1 Contribution Margin/cuốc (cùng giả định opex như `MARKET_PRICING_RESEARCH.md` Phần 6: VAT 10%, gateway 1.8%, opex cố định 880đ/cuốc — ASSUMPTION)

| Xe | Km | Khách trả | **Platform Profit** | **Biên (%)** |
|---|---|---|---|---|
| Bike | 5 | 25.530 | 3.976 | 15.6% |
| Bike | 20 | 79.620 | 12.738 | 16.0% |
| Bike | 80 | 253.980 | 40.985 | 16.1% |
| Car | 5 | 66.740 | 12.092 | 18.1% |
| Car | 20 | 193.560 | 32.637 | 16.9% |
| Car | 80 | 588.840 | 96.672 | 16.4% |
| XL | 5 | 91.600 | 16.119 | 17.6% |
| XL | 20 | 256.700 | 42.865 | 16.7% |
| XL | 80 | 771.100 | 126.198 | 16.4% |

Biên ổn định 15.6-18.1% ở mọi quãng đường/hạng xe — **đồng đều hơn** BRB hiện hành một chút (Phần 6 tài liệu trước cũng cho biên tương tự ~16-18%, vì cấu trúc opex/VAT tỷ lệ thuận với giá, không đổi theo cách tính Distance Fare). Khác biệt thật sự nằm ở **số tuyệt đối** cao hơn đáng kể do giá gốc đã hiệu chỉnh đúng thị trường.

### 17.2 Break-Even (company-wide, kế thừa phương pháp `MARKET_PRICING_RESEARCH.md` Phần 7.2)

Dùng cùng giả định chi phí cố định ~1 tỷ VND/tháng (ASSUMPTION, chưa CFO xác nhận): lợi nhuận trung bình/cuốc ở V3 (biên ~17%, quy mô trung bình ~250.000đ theo phân phối 312 kịch bản) ≈ **~42.500đ/cuốc** đóng góp cố định — cao hơn đáng kể so với ước lượng ~9.700đ/cuốc ở bảng giá cũ (`MARKET_PRICING_RESEARCH.md` Phần 7.2), nghĩa là số cuốc/tháng cần để hoà vốn công ty giảm từ ~103.000 xuống còn **~23.500 cuốc/tháng** (~780 cuốc/ngày) — một ngưỡng khả thi hơn nhiều ở giai đoạn tài xế còn ít (PRICING_STRATEGY §4, Giai đoạn 1-2).

### 17.3 Burn Rate

Không đổi phương pháp — Burn Rate (ECONOMY_ENGINE §10.2) đo tốc độ tiêu Promotion Fund + Incentive Fund/tháng, độc lập với việc hiệu chỉnh giá cước (Promotion/Incentive là ngân sách riêng, không phải một phần công thức giá — Phần 3). V3 không thay đổi cách đo Burn Rate, chỉ **tăng dư địa chịu đựng** của nó vì Contribution Margin/cuốc cao hơn (17.1) tạo nhiều "vốn đệm" hơn trước khi Burn Rate trở thành rủi ro.

---

## PHẦN 18 — SIMULATION (thiết kế, KHÔNG chạy)

### 18.1 Mục tiêu mở rộng từ 312 → 1.000 kịch bản

`MARKET_PRICING_RESEARCH.md` đã chạy 312 kịch bản thật (13 khoảng cách × 8 điều kiện × 3 hạng xe). Thiết kế mở rộng lên **1.000+** cho V3 (chỉ thiết kế lưới, không chạy ở tài liệu này):

| Chiều | V2 (đã chạy) | V3 (thiết kế, chưa chạy) |
|---|---|---|
| Khoảng cách | 13 mốc (2-60km) | 16 mốc (thêm 70/80/90/100km — vượt mốc trần giá dài, cần để kiểm tra Phần 13) |
| Điều kiện | 8 (Nội/Ngoại thành, Sân bay, Cao điểm, Mưa, Đêm, Cuối tuần, Lễ) | 8 (giữ nguyên) + **Emergency** (Phần 8.1) = 9 |
| Hạng xe | 3 (Bike/Car/XL) | 3 (giữ nguyên) |
| **Thành phố** — mới | 1 (ngầm định TP.HCM) | 5 (HCM/HN/ĐN/HP/CT — Phần 6) |
| **Promotion state** — mới | Không có trong lưới gốc (chỉ phân tích riêng ở sensitivity) | 3 trạng thái (Không áp dụng / Voucher trung bình 15% / Voucher sâu 30-50%) |

16 × 9 × 3 × 5 × 3 = **6.480 kịch bản khả dụng trong lưới đầy đủ** — vượt xa 1.000, nên đề xuất **lấy mẫu phân tầng (stratified sampling)** 1.000-1.200 kịch bản đại diện thay vì chạy toàn bộ lưới (giữ đúng tỷ trọng mỗi chiều, tránh thiên lệch về một thành phố/điều kiện), theo đúng phương pháp thống kê chuẩn cho việc mô phỏng lưới nhiều chiều khi chạy toàn bộ tổ hợp không cần thiết.

### 18.2 Tiêu chí đạt (Definition of Done cho lần chạy thật ở sprint tiếp theo — không phải kết quả ở đây)

- ≥ 1.000 kịch bản, 0 vi phạm an toàn (kế thừa nguyên bộ 5 ràng buộc đã kiểm chứng ở `PRICING_SIMULATION_REPORT.md` Phần 7: fare không âm, commission không âm, driver không âm tiền, platform không lỗ vô hạn, discount không vượt giá trị chuyến).
- Xác nhận: không kịch bản nào (kể cả 100km, Emergency, 5 thành phố) vi phạm Price Ceiling (Phần 13) sau khi có Distance Tier degressive.
- Đối chiếu Driver Income Index (Phần 2) ≥ 100% trên đa số kịch bản so với Market Reference.

---

## PHẦN 19 — CONFIG DESIGN (không hardcode)

### 19.1 Nguyên tắc

Kế thừa triệt để ECONOMY_ENGINE Phần 11 ("không hardcode toàn bộ Economy Engine") — áp dụng **lần đầu tiên cho chính công thức cước lõi** (hiện `fare.go`'s `DefaultFareConfig()` vẫn là hằng số Go, dù đã đúng VND). V3 tách 100% tham số ra khỏi code thành cấu hình.

### 19.2 Schema đề xuất (minh hoạ, YAML — không phải code thật, chỉ thiết kế cấu trúc dữ liệu)

```yaml
pricing_config:
  version: "v3.0.0"
  effective_date: "2026-XX-XX"   # bắt buộc — hỗ trợ "báo trước 30 ngày" BRB §1.4
  city: "hcm"                     # Phần 6 — 1 file/thành phố hoặc 1 bảng có cột city_id
  vehicle_classes:
    car:
      base_fare: 13000
      distance_tiers:             # Phần 4
        - {from_km: 0,  to_km: 2,  rate: 9500}
        - {from_km: 2,  to_km: 5,  rate: 8600}
        - {from_km: 5,  to_km: 10, rate: 7800}
        - {from_km: 10, to_km: 20, rate: 7000}
        - {from_km: 20, to_km: 40, rate: 6200}
        - {from_km: 40, to_km: 60, rate: 5400}
        - {from_km: 60, to_km: null, rate: 4600}
      traffic_time_per_min: 540    # Phần 5
      waiting_fee_per_min: 500
      waiting_grace_min: 3
      minimum_fare: 30000
      booking_fee: 3000
    xl: { ... }                   # cùng cấu trúc
    bike: { ... }
  airport:                        # Phần 7
    pickup_fee: {car: 15000, xl: 20000, bike: 0}
    dropoff_fee: {car: 5000, xl: 7000, bike: 0}
    queue_compensation_cap_min: 45
  surge:                          # Phần 8 — giữ nguyên bảng DSR đã có
    dsr_bands: [...]
    max_multiplier: 2.0
    emergency_override: false     # cờ vận hành, Operations bật/tắt qua Admin Portal, không qua release
  static_surcharge:
    night: 1.20
    holiday: 1.15
    rain: 1.15
    peak: 1.10
    static_cap: 1.60
  price_ceiling:                  # Phần 13
    standard_cap: 500000
    long_distance_cap: 1000000
    long_distance_threshold_km: 60
  city_coefficient:                # Phần 6
    hcm: 1.00
    hanoi: 1.00
    danang: 0.90
    haiphong: 0.85
    cantho: 0.82
  market_index:                    # Phần 14
    target_index: {car: 0.87, xl: 0.90, bike: 0.90}
    reference_market_fare_ref: "market_reference_2026q3"  # trỏ tới bảng riêng, cập nhật định kỳ
```

### 19.3 Nơi lưu — không phải quyết định kỹ thuật ở tài liệu business-level này, chỉ liệt kê lựa chọn

| Lựa chọn | Ưu điểm | Nhược điểm |
|---|---|---|
| File YAML/JSON trong repo, load lúc khởi động service | Đơn giản, dễ review qua PR, có Git history tự nhiên | Đổi giá vẫn cần deploy lại (không tuân thủ đầy đủ "effective-dated, không cần release" ECONOMY_ENGINE §11.2) |
| Bảng Postgres (`pricing_configs`, versioned + effective_date) | Đổi giá không cần release, hỗ trợ audit trail đầy đủ, hỗ trợ "hiệu lực tương lai" đúng yêu cầu BRB §1.4 | Cần thêm tầng cache để không query DB mỗi lần tính cước |
| Feature flag service (kiểu LaunchDarkly/tự xây) cho riêng `emergency_override` và `TargetIndex` | Bật/tắt tức thời, không cần cả deploy lẫn migration DB cho các cờ vận hành khẩn cấp | Không phù hợp cho toàn bộ `DistanceTiers` (quá nhiều tham số, không phải nhị phân) |

**Khuyến nghị (không tự quyết — CTO quyết định ở bước duyệt):** Postgres versioned cho `DistanceTiers`/`Airport`/`PriceCeiling`/`MarketIndex` (cần audit trail + effective-dated), feature flag riêng cho `emergency_override` (cần bật tức thời, không chờ 30 ngày vì đây là an toàn khẩn cấp, không phải thay đổi thương mại).

---

## PHẦN 20 — ROADMAP

| Giai đoạn | Nội dung | Điều kiện |
|---|---|---|
| **V3.1** | Distance Tier degressive (Phần 4) + Time Pricing 3 lớp (Phần 5) + Price Ceiling mới (Phần 13) — phần lõi giải quyết trực tiếp 2 vấn đề nghiêm trọng nhất (thu nhập tài xế quãng dài, vi phạm trần giá) | CPO/CFO duyệt bảng số Phần 4; chạy simulation thật (Phần 18) trước khi merge |
| **V3.2** | City Coefficient (Phần 6) + Airport 4 thành phần (Phần 7) + Stackable pairs cho Promotion (Phần 9.3) | V3.1 đã chạy ổn định ≥ 1 chu kỳ báo cáo tài chính (đề xuất: 1 tháng) tại TP.HCM |
| **V4** | Market Index tự động hoá (Phần 14 — cập nhật `ReferenceMarketFare` bán tự động thay vì khảo sát thủ công hoàn toàn), Emergency Override tích hợp cảnh báo thời tiết/thiên tai thật (thay vì cờ thủ công Operations), No Show Fee + Waiting Cap (Phần 3/5/13) | V3.2 ổn định, đa thành phố đã vận hành thật ≥ 2 thành phố |

---

## PHẦN 21 — MIGRATION PLAN (liệt kê, KHÔNG sửa ngay)

Chỉ thực hiện sau khi CPO/CFO/CTO phê duyệt chính thức từng giai đoạn ở Phần 20. Thứ tự bắt buộc (không đảo):

1. `docs/business/business-rule-bible-v1.0.md` — tu chính §2.2.1-§2.2.5 thành cấu trúc Distance Tier (không còn 1 số /km), thêm mục Airport 4 thành phần, thêm Price Ceiling dài hạn, thêm No Show Fee.
2. Quyết định lưu trữ config (Phần 19.3) — CTO chốt Postgres versioned vs YAML trước khi viết bất kỳ schema nào.
3. `backend/services/pricing/domain/entity/fare.go` — đổi `VehicleRates.PerKmRate` (1 field) thành `DistanceTiers []DistanceTier` (mảng) — đây là **breaking change cấu trúc dữ liệu**, không phải chỉ đổi giá trị.
4. `backend/services/pricing/app/fare_calculator.go` — viết lại vòng lặp cộng dồn theo bậc (Phần 4.3), thay vì 1 phép nhân.
5. `backend/services/pricing/simulation/pricing_constants.go` + toàn bộ file `simulation/*.go` — đồng bộ theo cấu trúc Tier mới, chạy lại simulation thật (Phần 18) trước khi đề xuất merge logic sang production (đúng nguyên tắc "chỉ merge khi simulation đã chứng minh tốt" đã áp dụng nhất quán từ sprint Dynamic Pricing Engine).
6. `backend/services/pricing/app/*_test.go`, `backend/services/pricing/grpc/handler_test.go` — cập nhật toàn bộ expected values theo cấu trúc mới (có tiền lệ từ 2 lần cập nhật trước trong dự án).
7. `backend/services/pricing/grpc/pricingpb/*.proto` — **cân nhắc mở rộng breakdown trả về** (Distance Fare hiện chỉ có 1 số, cần thêm optional trường "which tier(s) applied" nếu muốn hiển thị minh bạch cho rider — đây là thay đổi proto, cần CPO+CTO duyệt riêng vì ảnh hưởng API contract).
8. `apps/rider/lib/features/booking/domain/models/vehicle_option.dart` + `mock_booking_catalog.dart` + `mock_fare_calculator.dart` — cập nhật mock rate theo Distance Tier để giá ước tính trước khi đặt xe khớp giá thật.
9. `apps/rider/lib/features/booking/presentation/widgets/price_breakdown_sheet.dart` — cân nhắc hiển thị rõ bậc quãng đường nào đang áp dụng (minh bạch hoá, đúng BRB §1.2 Nguyên tắc 2), nếu Phần 21.7 (proto) được duyệt.
10. `backend/services/promotion/domain/entity/voucher.go` — nếu Phần 9.3 (stackable pairs) được duyệt: thêm bảng `stackable_pairs` mới, không sửa field `Stackable` hiện có (giữ tương thích ngược).
11. CHANGELOG.md — ghi nhận sau khi từng bước ở trên thực sự merge (không phải bây giờ).

---

## PHẦN 22 — RISK

| Loại | Rủi ro | Mức độ | Giảm thiểu |
|---|---|---|---|
| **Business** | Đắt hơn Grab ở chuyến < 12km (Phần 15) có thể làm rider nhạy cảm giá rời bỏ ở phân khúc chuyến ngắn nội thành (chiếm tỷ trọng lớn) | Cao | Theo dõi churn rate theo dải quãng đường trong 30 ngày đầu (Phần 23); sẵn sàng hạ bậc 1 ở V3.1 nếu dữ liệu xác nhận |
| **Business** | Market Reference (Phần 14) là dữ liệu khảo sát thủ công, có thể lệch giá thật tại một số khu vực/thời điểm (đã ghi nhận ở `MARKET_PRICING_RESEARCH.md` Phần 2.4) | Trung bình | Khảo sát lại định kỳ hàng quý (Phần 14.3), không dùng một lần rồi để cố định vô thời hạn |
| **Technical** | Đổi `PerKmRate` (1 field) thành `DistanceTiers` (mảng) là breaking change — mọi nơi đọc `FareBreakdown`/`VehicleRates` cần rà soát (Booking Service, Rider app mock) | Cao | Migration Plan (Phần 21) liệt kê tuần tự, chạy simulation trước khi đổi production (bước 5) |
| **Technical** | Config-driven (Phần 19) nếu chọn Postgres versioned cần thêm tầng cache — nếu thiếu cache, mỗi lần tính cước query DB có thể làm chậm response time Booking flow | Trung bình | CTO cần thiết kế cache invalidation khi `effective_date` tới hạn — nằm ngoài phạm vi tài liệu business-level này |
| **Driver** | Airport Pickup Fee mới (15.000đ Car) thấp hơn Airport Fee cũ + phần bù ẩn trong /km cũ cộng lại — cần đảm bảo tài xế sân bay không bị giảm thu nhập tuyệt đối so với hiện tại dù cấu trúc thay đổi | Trung bình | Đối chiếu Phần 16 (driver economics) trước và sau ở đúng kịch bản sân bay trước khi launch V3.1 |
| **Driver** | No Show Fee (Phần 3) mới — nếu ngưỡng thời gian quá ngắn, tài xế có thể lạm dụng huỷ sớm để "ăn" No Show Fee thay vì chờ khách thật | Trung bình | Cần Anti-Abuse review (đối chiếu ECONOMY_ENGINE Phần 9 — Collusion/Self-Ride detection) trước khi kích hoạt No Show Fee, không launch riêng lẻ mà không có cơ chế chống lạm dụng đi kèm |
| **Customer** | Airport Dropoff/Pickup phân biệt (Phần 7) có thể gây nhầm lẫn nếu không hiển thị rõ trong Price Breakdown Sheet | Thấp | Migration Plan bước 9 — cập nhật UI hiển thị minh bạch trước khi bật phí mới |
| **Customer** | Giá cao hơn Grab ở chuyến ngắn (rủi ro Business ở trên) đồng thời là rủi ro Customer trực tiếp — rider so giá 3 app trước khi đặt (đúng hành vi PS §1 ưu tiên #4 giả định) | Cao | Cùng giảm thiểu như dòng Business đầu tiên |

---

## PHẦN 23 — SUCCESS METRIC (sau launch)

| Mốc | KPI cần đạt | Hành động nếu không đạt |
|---|---|---|
| **7 ngày** | Không có kịch bản nào vi phạm Price Ceiling trong dữ liệu thật (Phần 13); Acceptance Rate không giảm so với baseline trước V3 | Rollback ngay lập tức nếu vi phạm Price Ceiling xảy ra thật (đây là ràng buộc an toàn cứng, không đợi 30/90 ngày) |
| **7 ngày** | Driver Income Index (Phần 2) tăng so với baseline BRB cũ, đo được ngay (không cần chờ hành vi rider thay đổi) | Nếu không tăng — kiểm tra lại Distance Tier có bị áp sai bậc trong code không (lỗi kỹ thuật khả dĩ nhất ở tuần đầu) |
| **30 ngày** | Customer Saving % (Phần 2) nằm trong dải mục tiêu (Bike/XL 8-12%, Car 10-15%) khi đối chiếu dữ liệu Market Reference cập nhật | Nếu lệch dải — điều chỉnh `TargetIndex` (Phần 14), không cần sửa `DistanceTiers` gốc |
| **30 ngày** | Churn rate theo dải quãng đường < 12km không tăng bất thường so với baseline (đối chiếu rủi ro Phần 22 "đắt hơn Grab ở chuyến ngắn") | Nếu tăng — cân nhắc hạ bậc 1 Distance Tier sớm hơn kế hoạch V3.1 gốc (Phần 15.3 khuyến nghị) |
| **90 ngày** | Repeat Rate ≥ 40% (Phần 2); Contribution Margin dương ổn định ≥ 95% chuyến (Phần 2/17) | Nếu Repeat Rate thấp — vấn đề có thể không nằm ở giá (đã hiệu chỉnh đúng thị trường), cần điều tra trải nghiệm sản phẩm khác, ngoài phạm vi tài liệu giá |
| **90 ngày** | Break-Even company-wide (Phần 17.2) tiến gần ngưỡng ~23.500 cuốc/tháng theo đúng tốc độ tăng trưởng tài xế dự kiến ở PRICING_STRATEGY §4 | Nếu lệch xa — CFO cần xác nhận lại giả định chi phí cố định (Phần 17.2 vốn là ASSUMPTION, chưa xác nhận thật) |

---

## GHI CHÚ PHƯƠNG PHÁP

- Mọi con số Distance Tier, Driver/Platform Economics (Phần 4, 15, 16, 17) được tính bằng script Python viết riêng cho tài liệu này (`v3_tiers.py`), tái sử dụng đúng công thức/giả định chi phí đã công bố ở `MARKET_PRICING_RESEARCH.md` (xăng, khấu hao, VAT, gateway, opex — tất cả vẫn là ASSUMPTION, chưa CFO xác nhận, nhắc lại ở đây để không đọc rời tài liệu mà quên giới hạn này).
- Rate card đối thủ (Phần 15) tái sử dụng nguyên số liệu đã nghiên cứu ở `MARKET_PRICING_RESEARCH.md` Phần 2 (WebSearch 2026-07-11, đã trích nguồn ở đó) — không nghiên cứu lại.
- Phần 18 (1.000 kịch bản) và mọi con số "chưa chạy" trong tài liệu này **là thiết kế lưới, không phải kết quả mô phỏng thật** — đúng yêu cầu nhiệm vụ "không chạy, chỉ thiết kế".
- Tài liệu này **không thay thế** BRB/PRICING_STRATEGY/ECONOMY_ENGINE/MARKET_PRICING_RESEARCH — mọi phát hiện/khuyến nghị của các tài liệu đó vẫn nguyên giá trị trừ khi được nêu rõ là thay đổi ở đây (ví dụ: Distance Tier thay thế trực tiếp bảng giá phẳng đề xuất tạm thời ở `MARKET_PRICING_RESEARCH.md` §8.2, nhưng giữ nguyên Base Fare/Minimum Fare/Booking Fee đã đề xuất ở đó).

---

*Kết thúc tài liệu — Panda Pricing V3 Design — v0.1 (Thiết kế kiến trúc).*
*Không sửa Business Rule Bible. Không sửa Pricing Service. Không sửa Promotion Engine. Không sửa UI. Không build. Không commit.*
*Tài liệu này phải được CPO, CFO, CTO phê duyệt theo đúng thứ tự Roadmap (Phần 20) và Migration Plan (Phần 21) trước khi bất kỳ dòng code nào được viết.*
