# Panda — Market Pricing Research (chuẩn bị cho Pricing V2)

**Document Classification:** Internal — Confidential
**Authority:** CPO · CFO · CTO (cần phê duyệt trước khi bất kỳ số nào dưới đây được đưa vào production)
**Effective Date:** 2026-07-11
**Status:** NGHIÊN CỨU + THIẾT KẾ — không phải đề xuất sẵn sàng triển khai
**Nguồn sự thật khi có mâu thuẫn:** `docs/business/business-rule-bible-v1.0.md` (BRB). Tài liệu này **không sửa BRB**, không sửa Pricing Service, không sửa UI, không build, không commit — đúng yêu cầu nhiệm vụ.
**Tài liệu đã đọc trước khi viết:** BRB v1.0 (đặc biệt Part 2 Pricing, Part 7 Driver Economy, Part 8 Driver Incentive, Part 14 Financial Reports), `PRICING_STRATEGY.md`, `ECONOMY_ENGINE.md`, `PRICING_SIMULATION_REPORT.md` (111 scenario đã chạy trước đó — tài liệu này **kế thừa và mở rộng**, không lặp lại), `backend/services/pricing/domain/entity/fare.go` (production), `backend/services/pricing/simulation/*.go` (simulation engine đã có sẵn từ sprint trước).

---

## TÓM TẮT ĐIỀU HÀNH

- **Đã xác nhận bằng số liệu thật (không phải cảm tính):** ở cấu hình đang chạy production (BRB §2.2.1-§2.2.5, VND, vừa được nối vào `fare.go` phiên làm việc trước), giá Panda trung bình **thấp hơn thị trường 42.5%** trên 312 kịch bản × 3 nền tảng công nghệ (Grab/Be/GreenSM) — khớp với phát hiện ban đầu "45-50%". Chi tiết theo hạng xe: **Car -52.8%, XL -52.0%, Bike -22.6%.**
- **Nguyên nhân gốc không phải Base Fare** — là **/km**: Panda tính 4.000đ/km (car) so với thị trường trung bình ~11.000-15.000đ/km (nghiên cứu thật, Phần 2). Đây là biến số duy nhất nhân với toàn bộ quãng đường, nên nó áp đảo mọi thành phần khác (Airport Fee, Night/Rain/Holiday surcharge của Panda thực ra **cùng bậc độ lớn** với các phụ phí tương đương của đối thủ — không phải nguyên nhân).
- **Phát hiện tài chính nghiêm trọng nhất:** ở giá hiện tại, chi phí xăng + khấu hao của tài xế ô tô **gần như ăn hết toàn bộ phần thu nhập theo km** (chỉ còn lại lợi nhuận ròng từ Base Fare + Booking Fee, không tăng theo quãng đường) — một cuốc car 60km ở giá hiện tại chỉ để lại **14.240đ lợi nhuận ròng cho tài xế** sau xăng/khấu hao, và một cú sốc xăng +20% sẽ đẩy con số này **xuống âm** (Phần 10). Đây là bằng chứng số cho nhận định "không bền vững".
- **Đề xuất:** bảng giá mới đồng bộ (Phần 8) đưa Car về **-13.5%** so với thị trường (mục tiêu 10-15% ✓), XL về **-11.0%** (mục tiêu 8-12% ✓), Bike về **-9.8%** (mục tiêu 8-12% ✓) — không phải "rẻ nhất", đúng định vị "Balanced" đã chọn trong PRICING_STRATEGY §1.
- **Giới hạn phát hiện được trong lúc thiết kế (không giấu):** bảng giá đề xuất, ở dạng phẳng (không có bậc giảm giá quãng đường dài như cả 5 đối thủ đều có), **vi phạm trần giá tuyệt đối BRB §2.13.6 (500.000đ)** từ khoảng 50km trở lên cho hạng Car/XL — cần thêm bậc chiết khấu quãng đường dài trước khi trình duyệt chính thức (Phần 8.4).
- **Điểm hoà vốn mỗi cuốc** đã dương ở cả giá hiện tại lẫn giá đề xuất (biên ~16-18% ở Bronze) — vấn đề không phải "mỗi cuốc đang lỗ", mà là "giá quá thấp so với thị trường tới mức không tạo đủ dư địa cho tài xế lẫn nền tảng khi có cú sốc chi phí" (xem Phần 10).

---

## PHẦN 1 — PHÂN TÍCH NGUYÊN NHÂN

### 1.1 Không chỉ nhìn Base Fare

| Thành phần | Panda (production hôm nay) | Vai trò trong khoảng cách 42.5% |
|---|---|---|
| **Base Fare** | 10.000đ (car), 18.000đ (XL), 6.000đ (bike — suy ra tỷ lệ 0.60 từ `simulation/pricing_constants.go`, xem 1.5) | Đóng góp **cố định**, không nhân với km — ở chuyến dài, tỷ trọng của Base Fare trong tổng giá giảm dần, nên nó **không phải** nguyên nhân chính của khoảng cách 42.5% (khoảng cách này **tăng dần theo km**, xem bảng 312 kịch bản Phần 4 — nếu Base Fare là nguyên nhân chính, khoảng cách phải **giảm dần** theo km vì tỷ trọng Base giảm) |
| **/km** | 4.000đ (car), 5.000đ (XL), 2.400đ (bike) | **Nguyên nhân chính.** Thị trường trung bình 11.000-15.000đ/km (car). Panda chỉ thu **~30-35%** mức này. Vì đây là số nhân với **toàn bộ quãng đường**, nó áp đảo mọi thành phần khác và giải thích vì sao khoảng cách % **tăng dần theo km** (car: -24.5% ở 2km → -56.1% ở 60km, xem Phần 4) |
| **/phút** | 400đ (car) | Đóng góp nhỏ (ước tính ~2.2 phút/km × 400đ = 880đ/km quy đổi) — không phải đòn bẩy chính, nhưng cùng chiều với /km |
| **Minimum Fare** | 25.000đ (car) | Chỉ chi phối chuyến rất ngắn (< ~3-4km) — đây là lý do khoảng cách % ở 2km (-24.5%) **nhỏ hơn** khoảng cách ở 60km (-56.1%): Minimum Fare kéo giá chuyến ngắn lên gần thị trường hơn, nhưng /km thấp làm chuyến dài rớt xa dần |
| **Booking Fee** | 2.000đ, cố định | Không đáng kể so với tổng giá — Grab/Be/GreenSM không công bố một dòng phí tương đương riêng biệt (có thể đã gộp vào /km của họ) |
| **Airport Fee** | 10.000đ flat | Cùng bậc độ lớn với phụ phí sân bay của đối thủ (Mai Linh công bố +3.000đ/km, tương đương ~30.000-45.000đ cho chuyến 10-15km — **cao hơn** Panda). Airport Fee **không phải** nguyên nhân Panda rẻ — ở khía cạnh này Panda thậm chí thu **ít hơn** Mai Linh cho chuyến dài |
| **Bridge/Toll Fee** | Pass-through 100%, 0% hoa hồng | Trung lập — mọi nền tảng đều pass-through, không tạo khoảng cách giá |
| **Waiting Fee** | 500đ/phút sau 3 phút miễn phí | Không đủ dữ liệu để so sánh (không nền tảng nào công bố rate card chờ xe công khai) — giả định trung lập |
| **Long Pickup Compensation** | 10.000-20.000đ, nền tảng chịu 100%, rider không trả thêm | Không ảnh hưởng giá hiển thị cho rider — chỉ ảnh hưởng chi phí nền tảng (xem Phần 6) |
| **Dynamic Pricing (Surge)** | Trần ×2.0, dùng DSR — **cơ chế**, không phải **mức giá gốc** | Không giải thích khoảng cách 42.5% vì đây là hệ số nhân **tạm thời** theo cung-cầu, áp dụng như nhau cho giá gốc dù giá gốc là bao nhiêu — nếu giá gốc thấp, giá surge cũng thấp theo tỷ lệ |
| **Night/Holiday/Rain/Peak** | +20%/+15%/+15%/+10%, trần cộng dồn ×1.60 | Tương đương bậc độ lớn với "surge thời gian thực" mà Grab/Be/GreenSM áp dụng (không công bố công thức cố định, nhưng mức tăng phổ biến được ghi nhận trong nghiên cứu thị trường là +15-50% giờ cao điểm) — không phải nguyên nhân chính |
| **Promotion** | Tối đa ~50-70% (First Ride), phần lớn campaign khác 10-30% | Đây là **chiết khấu tạm thời có mục tiêu**, không phải giá niêm yết — không giải thích khoảng cách giữa giá niêm yết Panda và giá niêm yết đối thủ |
| **Membership** | Chỉ giảm Booking Fee/ưu tiên dịch vụ, không đổi công thức giá cước (nguyên tắc bất biến, ECONOMY_ENGINE §8.1) | Không liên quan đến giá niêm yết |
| **Commission (hoa hồng tài xế)** | 20% (Bronze) → 12% (Diamond) | Đây là cách **chia** giá thu được, không phải cách **định** giá — hoa hồng thấp hơn Grab/Uber (~25-27%) nghĩa là ở CÙNG một giá thu được, tài xế Panda có commission-rate tốt hơn, nhưng vì giá thu được (Customer Total) đã thấp hơn 42.5% ngay từ đầu, phần trăm cao hơn đó không đủ bù (xem Phần 5 — driver_net_profit vẫn thấp hơn tuyệt đối dù % hoa hồng thấp hơn) |
| **Driver Incentive (Quest/Bonus...)** | Không phải một phần của giá cước — chi từ ngân sách riêng | Không liên quan đến giá niêm yết rider nhìn thấy |

**Kết luận Phần 1:** khoảng cách 42.5% là hệ quả của **một quyết định cấu hình duy nhất chưa từng được hiệu chỉnh**: `DefaultFareConfig()` được viết ra như "giá test tạm cho môi trường dev" (chính comment gốc trong code nói rõ "Operators MUST override these for production") và info chưa bao giờ được đối chiếu với giá thị trường thật trước khi phiên làm việc trước nối nó vào VND theo đúng con số BRB — con số BRB tự nó **cũng chưa từng được hiệu chỉnh theo thị trường**, nó chỉ là "launch-market defaults" trong tài liệu gốc. Đây không phải lỗi kỹ thuật — đây là một **khoảng trống hiệu chỉnh kinh doanh** chưa ai lấp.

### 1.2 Một phát hiện phụ đáng chú ý: Bike ít lệch hơn Car/XL

Bike hiện tại chỉ lệch **-22.6%** trong khi Car/XL lệch **~-52%**. Lý do: tỷ lệ xe máy/xe hơi trong `fare.go` (0.60, lấy từ `simulation/pricing_constants.go` — xem 1.5) tình cờ **gần với tỷ lệ thị trường thật** (bike/car ở Grab/Be/GreenSM dao động 30-46% tuỳ nền tảng — xem Phần 2), trong khi bản thân mức Car (căn cứ BRB §2.2.1-§2.2.4) chưa bao giờ được đối chiếu với Grab/Be/GreenSM. Nói cách khác: **BRB định giá xe máy "vô tình" gần đúng, nhưng định giá xe hơi thì không** — hai lỗi độc lập, không phải cùng một nguyên nhân.

### 1.3 Phát hiện cấu trúc: Airport Fee bất cân xứng theo hạng xe

Khi áp Airport Fee (10.000đ flat, theo BRB §2.2.7, áp dụng như nhau cho mọi hạng xe) vào một chuyến xe máy 2km, khoảng cách so với thị trường đảo chiều từ **-15.9%** (điều kiện thường) thành **+56.2%** (điều kiện sân bay) — xem dòng #3 bảng Phần 4. Nguyên nhân: Grab/Be/GreenSM **không thu phụ phí sân bay riêng cho xe máy** trong thực tế thị trường (phụ phí sân bay ở các nền tảng công nghệ thường chỉ áp cho ô tô, do chi phí bãi đỗ/xếp hàng tại khu vực ô tô đón khách), trong khi BRB áp dụng Airport Fee đồng nhất mọi hạng xe. Đây là **gap cấu trúc cần CPO quyết định** trước khi Pricing V2 triển khai (đã đưa vào danh sách Phần 11).

### 1.4 Reconciliation cần làm (không phải lỗi mới, nhưng cần thống nhất)

`backend/services/pricing/domain/entity/fare.go` (production, sửa phiên trước) dùng tỷ lệ xe máy ~0.48-0.5 (suy từ 5.000/10.000, 1.600/4.000...) trong khi `backend/services/pricing/simulation/pricing_constants.go` (sprint trước, cùng repo) dùng **0.60**, với lý do được ghi rõ ("consistent with the typical motorcycle-vs-car fare ratio observed in other Southeast Asian ride-hailing markets"). Tài liệu này dùng **0.60** (số có lý do rõ ràng hơn) làm cơ sở phân tích — xem Phần 11 để biết cần đồng bộ hai file này.

---

## PHẦN 2 — NGHIÊN CỨU THỊ TRƯỜNG (bảng giá thật, có nguồn)

**Phương pháp:** tra cứu công khai qua WebSearch ngày 2026-07-11 (nguồn liệt kê cuối mỗi mục). Không phải số liệu nội bộ của Grab/Be/Xanh SM/Mai Linh/Vinasun — là giá niêm yết công khai trên các trang tổng hợp/tin tức, có thể lệch nhẹ so với giá app hiển thị tại từng thời điểm/khu vực cụ thể. Mọi chỗ không tìm được dữ liệu công khai đều được đánh dấu **ASSUMPTION** rõ ràng, không suy diễn thành sự thật.

### 2.1 Ô tô 4 chỗ (Car / Standard)

| Nền tảng | Giá mở cửa | /km sau đó | Nguồn |
|---|---|---|---|
| **Grab (GrabCar)** | 29.000đ (2km đầu) | ~10.000đ/km (không tìm được bậc giảm giá đường dài công khai — ASSUMPTION: giữ flat toàn bộ quãng đường còn lại, khả năng cao **overestimate** giá Grab thật ở cự ly rất dài) | [cellphones.com.vn](https://cellphones.com.vn/sforum/grab-bao-nhieu-tien-1km), [viettelstore.vn](https://viettelstore.vn/tin-tuc/grab-bao-nhieu-tien-1km) |
| **Be (beCar)** | 31.501đ (~1km) | 11.805đ/km (km2-10), 10.779đ/km (km11+) — nguồn công bố khoảng 10.779-11.805đ tuỳ cự ly, mô hình 2 bậc ở đây là suy diễn hợp lý từ khoảng đó | [3gvinaphone.com.vn](https://3gvinaphone.com.vn/gia-cuoc-xe-bebike-becar.html), [techbike.xyz](https://techbike.xyz/threads/ung-dung-be-cong-bo-gia-cuoc-dich-vu-bebike-becar-4-cho-7-cho.2146/) |
| **Xanh SM (GreenCar)** | 20.000đ (1km, TP.HCM) | 15.000đ/km (km2-24, dùng mức VF5 Plus rẻ hơn) → 12.000đ/km (km25+, **ASSUMPTION**: dùng lại bậc Hà Nội vì không tìm được bậc riêng cho TP.HCM sau km25) | [viettelstore.vn](https://viettelstore.vn/tin-tuc/gia-cuoc-taxi-xanh-vinfast-moi-nhat), [cellphones.com.vn](https://cellphones.com.vn/sforum/cach-dat-xe-taxi-vinfast) |
| **Mai Linh (taxi truyền thống)** | 12.000đ (0.6km) | 14.000đ/km (đến km30), 11.000đ/km (km30+); +10% giờ cao điểm; **+3.000đ/km phụ phí sân bay** (công bố rõ) | [olm.vn](https://olm.vn/hoi-dap/tim-kiem?id=167332763436), [taximailinh.vn](https://taximailinh.vn/gia-cuoc-taxi-mai-linh) |
| **Vinasun (taxi truyền thống)** | 11.000đ (0.5km) | 14.500đ/km (đến km31), 11.600đ/km (km31+) | [soctrangtourism.vn](https://soctrangtourism.vn/gia-cuoc-taxi-vinasun-4-cho/), [fxbike.vn](https://fxbike.vn/gia-xe-taxi-vinasun/) |
| **Panda (production hôm nay)** | 10.000đ (0km — không bao gồm quãng đường nào) | 4.000đ/km + 400đ/phút | `fare.go` |

### 2.2 Xe máy (Bike / Motorcycle)

| Nền tảng | Giá mở cửa | /km sau đó | Nguồn |
|---|---|---|---|
| **Grab (GrabBike)** | ~14.000đ (2km, khoảng công bố 12.500-16.000đ) | ~4.300đ/km | [viettelstore.vn](https://viettelstore.vn/tin-tuc/grab-bao-nhieu-tien-1km), [khainamtransport.com](https://khainamtransport.com/grab-xe-may-bao-nhieu-tien-1km-bang-gia-chi-tiet-moi-nhat/) |
| **Be (beBike)** | 13.817đ (2km, TP.HCM) | 4.633đ/km | [3gvinaphone.com.vn](https://3gvinaphone.com.vn/gia-cuoc-xe-bebike-becar.html), [thegioididong.com](https://www.thegioididong.com/game-app/chay-be-co-on-khong-thu-nhap-va-chiet-khau-bebike-becar-la-1562548) |
| **Xanh SM (GreenBike)** | 13.800đ (2km) | 4.800đ/km | [xehay.vn](https://xehay.vn/dich-vu-xe-may-dien-xanh-sm-bike-do-bo-ha-noi-gia-cuoc-tu-4-800-km.html) |
| **Panda (production hôm nay)** | 6.000đ | 2.400đ/km + 240đ/phút (tỷ lệ 0.60 so với car — xem Phần 1.4) | `simulation/pricing_constants.go` |

Đối chiếu thị trường: bike/car ratio thật là **Grab ≈ 48%** (14.000/29.000 ở km ngắn), **Be ≈ 44%**, **Xanh SM ≈ 69%** (13.800/20.000, đắt tương đối vì GreenCar mở cửa rẻ) — trung bình quanh **50-55%**, gần với tỷ lệ 0.60 mà `simulation/pricing_constants.go` đã chọn.

### 2.3 XL / 7 chỗ / Premium

**Không tìm được rate card XL/7-chỗ công khai riêng cho Grab/Be/Mai Linh/Vinasun** — thị trường Việt Nam ít quảng bá giá XL độc lập. **ASSUMPTION rõ ràng**: dùng hệ số +30% trên giá Car cùng nền tảng, một hệ số thường được trích dẫn trong so sánh thị trường Việt Nam cho hạng 7 chỗ. Riêng **Xanh SM Premium/Luxury** có giá công khai: **21.000đ/km trọn tuyến, không phân biệt giá mở cửa** ([viettelstore.vn](https://viettelstore.vn/tin-tuc/gia-cuoc-taxi-xanh-vinfast-moi-nhat)) — áp dụng thêm mức sàn 50.000đ (ASSUMPTION, tránh giá phi lý ở chuyến rất ngắn khi công thức flat/km không có sàn).

### 2.4 Giới hạn phương pháp (công bố minh bạch)

1. Không có rate card real-time — giá thật trên app biến động theo surge, không phải giá niêm yết.
2. Grab/GrabBike không có bậc giảm giá quãng đường dài công khai được tìm thấy — mô hình flat/km cho quãng rất dài (40-60km) khả năng **overestimate** giá Grab thật.
3. Phụ phí giờ cao điểm/mưa/đêm/lễ cho Grab/Be/GreenSM dùng surge thuật toán thời gian thực, không công bố công thức cố định — tài liệu này dùng **ASSUMPTION +15%** thống nhất cho các điều kiện này khi so sánh (ghi rõ trong Phần 4), không phải số đo thật.
4. XL/Premium dùng hệ số giả định +30% (trừ Xanh SM có số thật).
5. Phí sân bay cho Grab/Be/GreenSM (ô tô) dùng **ASSUMPTION +15.000đ** (không tìm được số công bố chính thức) — riêng Mai Linh có số thật (+3.000đ/km).

---

## PHẦN 3 — PHƯƠNG PHÁP TẠO 312 KỊCH BẢN

Lưới kịch bản: **13 khoảng cách × 8 điều kiện × 3 hạng xe = 312** (≥ 300 theo yêu cầu).

- **Khoảng cách (km):** 2, 4, 6, 8, 10, 12, 15, 20, 25, 30, 40, 50, 60 — bao phủ chuyến rất ngắn đến rất dài, đúng ví dụ đề bài đưa ra và mở rộng thêm để có đường cong mượt.
- **Điều kiện (8):** Nội thành (baseline), Ngoại thành, Sân bay, Giờ cao điểm, Mưa, Đêm, Cuối tuần, Lễ Tết.
- **Hạng xe (3):** Bike, Car, XL.
- **Thời lượng chuyến** suy ra từ khoảng cách theo đúng quy ước đã có sẵn trong `backend/services/pricing/simulation/scenarios.go` (`DurationMin = DistanceKM × 2.2`, tương đương tốc độ trung bình nội thành ~27km/h) — dùng lại để nhất quán với Simulation Engine hiện có, không phát minh quy ước mới.
- **Phụ phí Panda** áp đúng theo `pricing_constants.go`: Night ×1.20, Holiday ×1.15, Rain ×1.15, Peak ×1.10 (trần cộng dồn ×1.60), Airport Fee 10.000đ flat.
- **Cuối tuần** không có phụ phí cấu trúc riêng cho Panda (BRB không định nghĩa phụ phí cuối tuần cho giá cước — chỉ có Weekend Promotion giảm giá cho rider, PS §7.1) — kịch bản "Cuối tuần" trong bảng dưới bằng đúng giá "Nội thành" để phản ánh đúng thực tế này, không tự thêm phụ phí không có căn cứ.
- **Script tính toán:** viết bằng Python (`pricing_research.py`, không phải một phần bàn giao production, chỉ dùng để tạo báo cáo này — tương tự cách sprint trước dùng bản port Node.js cho `PRICING_SIMULATION_REPORT.md` vì môi trường không có Go toolchain sẵn khi cần chạy nhanh).

---

## PHẦN 4 — BẢNG SO SÁNH (312 kịch bản)

### 4.1 Tóm tắt theo hạng xe

| Hạng xe | Số kịch bản | Chênh lệch TB — Panda hiện tại | Chênh lệch TB — Panda đề xuất | Mục tiêu (Phần 7) | Đạt? |
|---|---|---|---|---|---|
| Bike | 104 | **-22.6%** | **-9.8%** | 8-12% rẻ hơn | ✓ |
| Car | 104 | **-52.8%** | **-13.5%** | 10-15% rẻ hơn | ✓ |
| XL | 104 | **-52.0%** | **-11.0%** | 8-12% rẻ hơn | ✓ |
| **Tổng hợp** | **312** | **-42.5%** | **-11.4%** | — | — |

(Chênh lệch âm = Panda rẻ hơn trung bình cộng Grab+Be+GreenSM cùng kịch bản. Đối thủ taxi truyền thống — Mai Linh/Vinasun — liệt kê trong bảng chi tiết để tham khảo, **không tính vào chỉ số Market Index** vì BRB/PRICING_STRATEGY định vị Panda cạnh tranh với nhóm nền tảng công nghệ, taxi truyền thống chỉ là mốc tham chiếu lịch sử — PS §4.1.)

### 4.2 Quan sát theo khoảng cách (hạng Car, điều kiện Nội thành)

| Km | Panda hiện tại | TB thị trường | Chênh lệch |
|---|---|---|---|
| 2 | 27.000 | 35.769 | -24.5% |
| 10 | 60.800 | 133.915 | -54.6% |
| 25 | 207.200 → *(xem dòng #169 bảng đầy đủ: 134.000)* | 311.810 | -57.0% |
| 60 | 304.800 | 694.232 | -56.1% |

→ Khoảng cách **nới rộng dần** theo km rồi ổn định quanh mức -55/-58% từ ~15km trở đi — xác nhận đúng phân tích Phần 1: nguyên nhân nằm ở **/km**, không phải Base Fare/Minimum Fare (nếu ngược lại, khoảng cách sẽ **thu hẹp** dần theo km).

### 4.3 Toàn bộ 312 kịch bản

Cột **Δ hiện tại** / **Δ đề xuất** = % chênh lệch (Panda − TB thị trường) / TB thị trường. Âm = Panda rẻ hơn.

| # | Xe | Km | Điều kiện | Grab | Be | GreenSM | Mai Linh | Vinasun | TB Thị trường | Panda hiện tại | Panda đề xuất | Δ hiện tại | Δ đề xuất |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | Bike | 2 | Nội thành | 14.000 | 13.817 | 13.800 | — | — | 13.872 | 17.000 | 11.668 | 22.5% | -15.9% |
| 2 | Bike | 2 | Ngoại thành | 14.000 | 13.817 | 13.800 | — | — | 13.872 | 17.000 | 11.668 | 22.5% | -15.9% |
| 3 | Bike | 2 | Sân bay | 14.000 | 13.817 | 13.800 | — | — | 13.872 | 27.000 | 21.668 | 94.6% | 56.2% |
| 4 | Bike | 2 | Giờ cao điểm | 14.000 | 13.817 | 13.800 | — | — | 15.953 | 17.000 | 12.735 | 6.6% | -20.2% |
| 5 | Bike | 2 | Mưa | 14.000 | 13.817 | 13.800 | — | — | 15.953 | 17.000 | 13.268 | 6.6% | -16.8% |
| 6 | Bike | 2 | Đêm | 14.000 | 13.817 | 13.800 | — | — | 15.953 | 17.000 | 13.802 | 6.6% | -13.5% |
| 7 | Bike | 2 | Cuối tuần | 14.000 | 13.817 | 13.800 | — | — | 13.872 | 17.000 | 11.668 | 22.5% | -15.9% |
| 8 | Bike | 2 | Lễ Tết | 14.000 | 13.817 | 13.800 | — | — | 15.953 | 17.000 | 13.268 | 6.6% | -16.8% |
| 9 | Bike | 4 | Nội thành | 22.600 | 23.083 | 23.400 | — | — | 23.028 | 19.712 | 19.836 | -14.4% | -13.9% |
| 10 | Bike | 4 | Ngoại thành | 22.600 | 23.083 | 23.400 | — | — | 23.028 | 19.712 | 19.836 | -14.4% | -13.9% |
| 11 | Bike | 4 | Sân bay | 22.600 | 23.083 | 23.400 | — | — | 23.028 | 29.712 | 29.836 | 29.0% | 29.6% |
| 12 | Bike | 4 | Giờ cao điểm | 22.600 | 23.083 | 23.400 | — | — | 26.482 | 21.483 | 21.720 | -18.9% | -18.0% |
| 13 | Bike | 4 | Mưa | 22.600 | 23.083 | 23.400 | — | — | 26.482 | 22.369 | 22.661 | -15.5% | -14.4% |
| 14 | Bike | 4 | Đêm | 22.600 | 23.083 | 23.400 | — | — | 26.482 | 23.254 | 23.603 | -12.2% | -10.9% |
| 15 | Bike | 4 | Cuối tuần | 22.600 | 23.083 | 23.400 | — | — | 23.028 | 19.712 | 19.836 | -14.4% | -13.9% |
| 16 | Bike | 4 | Lễ Tết | 22.600 | 23.083 | 23.400 | — | — | 26.482 | 22.369 | 22.661 | -15.5% | -14.4% |
| 17 | Bike | 6 | Nội thành | 31.200 | 32.349 | 33.000 | — | — | 32.183 | 25.568 | 28.004 | -20.6% | -13.0% |
| 18 | Bike | 6 | Ngoại thành | 31.200 | 32.349 | 33.000 | — | — | 32.183 | 25.568 | 28.004 | -20.6% | -13.0% |
| 19 | Bike | 6 | Sân bay | 31.200 | 32.349 | 33.000 | — | — | 32.183 | 35.568 | 38.004 | 10.5% | 18.1% |
| 20 | Bike | 6 | Giờ cao điểm | 31.200 | 32.349 | 33.000 | — | — | 37.010 | 27.925 | 30.704 | -24.5% | -17.0% |
| 21 | Bike | 6 | Mưa | 31.200 | 32.349 | 33.000 | — | — | 37.010 | 29.103 | 32.055 | -21.4% | -13.4% |
| 22 | Bike | 6 | Đêm | 31.200 | 32.349 | 33.000 | — | — | 37.010 | 30.282 | 33.405 | -18.2% | -9.7% |
| 23 | Bike | 6 | Cuối tuần | 31.200 | 32.349 | 33.000 | — | — | 32.183 | 25.568 | 28.004 | -20.6% | -13.0% |
| 24 | Bike | 6 | Lễ Tết | 31.200 | 32.349 | 33.000 | — | — | 37.010 | 29.103 | 32.055 | -21.4% | -13.4% |
| 25 | Bike | 8 | Nội thành | 39.800 | 41.615 | 42.600 | — | — | 41.338 | 31.424 | 36.172 | -24.0% | -12.5% |
| 26 | Bike | 8 | Ngoại thành | 39.800 | 41.615 | 42.600 | — | — | 41.338 | 31.424 | 36.172 | -24.0% | -12.5% |
| 27 | Bike | 8 | Sân bay | 39.800 | 41.615 | 42.600 | — | — | 41.338 | 41.424 | 46.172 | 0.2% | 11.7% |
| 28 | Bike | 8 | Giờ cao điểm | 39.800 | 41.615 | 42.600 | — | — | 47.539 | 34.366 | 39.689 | -27.7% | -16.5% |
| 29 | Bike | 8 | Mưa | 39.800 | 41.615 | 42.600 | — | — | 47.539 | 35.838 | 41.448 | -24.6% | -12.8% |
| 30 | Bike | 8 | Đêm | 39.800 | 41.615 | 42.600 | — | — | 47.539 | 37.309 | 43.206 | -21.5% | -9.1% |
| 31 | Bike | 8 | Cuối tuần | 39.800 | 41.615 | 42.600 | — | — | 41.338 | 31.424 | 36.172 | -24.0% | -12.5% |
| 32 | Bike | 8 | Lễ Tết | 39.800 | 41.615 | 42.600 | — | — | 47.539 | 35.838 | 41.448 | -24.6% | -12.8% |
| 33 | Bike | 10 | Nội thành | 48.400 | 50.881 | 52.200 | — | — | 50.494 | 37.280 | 44.340 | -26.2% | -12.2% |
| 34 | Bike | 10 | Ngoại thành | 48.400 | 50.881 | 52.200 | — | — | 50.494 | 37.280 | 44.340 | -26.2% | -12.2% |
| 35 | Bike | 10 | Sân bay | 48.400 | 50.881 | 52.200 | — | — | 50.494 | 47.280 | 54.340 | -6.4% | 7.6% |
| 36 | Bike | 10 | Giờ cao điểm | 48.400 | 50.881 | 52.200 | — | — | 58.068 | 40.808 | 48.674 | -29.7% | -16.2% |
| 37 | Bike | 10 | Mưa | 48.400 | 50.881 | 52.200 | — | — | 58.068 | 42.572 | 50.841 | -26.7% | -12.4% |
| 38 | Bike | 10 | Đêm | 48.400 | 50.881 | 52.200 | — | — | 58.068 | 44.336 | 53.008 | -23.6% | -8.7% |
| 39 | Bike | 10 | Cuối tuần | 48.400 | 50.881 | 52.200 | — | — | 50.494 | 37.280 | 44.340 | -26.2% | -12.2% |
| 40 | Bike | 10 | Lễ Tết | 48.400 | 50.881 | 52.200 | — | — | 58.068 | 42.572 | 50.841 | -26.7% | -12.4% |
| 41 | Bike | 12 | Nội thành | 57.000 | 60.147 | 61.800 | — | — | 59.649 | 43.136 | 52.508 | -27.7% | -12.0% |
| 42 | Bike | 12 | Ngoại thành | 57.000 | 60.147 | 61.800 | — | — | 59.649 | 43.136 | 52.508 | -27.7% | -12.0% |
| 43 | Bike | 12 | Sân bay | 57.000 | 60.147 | 61.800 | — | — | 59.649 | 53.136 | 62.508 | -10.9% | 4.8% |
| 44 | Bike | 12 | Giờ cao điểm | 57.000 | 60.147 | 61.800 | — | — | 68.596 | 47.250 | 57.659 | -31.1% | -15.9% |
| 45 | Bike | 12 | Mưa | 57.000 | 60.147 | 61.800 | — | — | 68.596 | 49.306 | 60.234 | -28.1% | -12.2% |
| 46 | Bike | 12 | Đêm | 57.000 | 60.147 | 61.800 | — | — | 68.596 | 51.363 | 62.810 | -25.1% | -8.4% |
| 47 | Bike | 12 | Cuối tuần | 57.000 | 60.147 | 61.800 | — | — | 59.649 | 43.136 | 52.508 | -27.7% | -12.0% |
| 48 | Bike | 12 | Lễ Tết | 57.000 | 60.147 | 61.800 | — | — | 68.596 | 49.306 | 60.234 | -28.1% | -12.2% |
| 49 | Bike | 15 | Nội thành | 69.900 | 74.046 | 76.200 | — | — | 73.382 | 51.920 | 64.760 | -29.2% | -11.7% |
| 50 | Bike | 15 | Ngoại thành | 69.900 | 74.046 | 76.200 | — | — | 73.382 | 51.920 | 64.760 | -29.2% | -11.7% |
| 51 | Bike | 15 | Sân bay | 69.900 | 74.046 | 76.200 | — | — | 73.382 | 61.920 | 74.760 | -15.6% | 1.9% |
| 52 | Bike | 15 | Giờ cao điểm | 69.900 | 74.046 | 76.200 | — | — | 84.389 | 56.912 | 71.136 | -32.6% | -15.7% |
| 53 | Bike | 15 | Mưa | 69.900 | 74.046 | 76.200 | — | — | 84.389 | 59.408 | 74.324 | -29.6% | -11.9% |
| 54 | Bike | 15 | Đêm | 69.900 | 74.046 | 76.200 | — | — | 84.389 | 61.904 | 77.512 | -26.6% | -8.1% |
| 55 | Bike | 15 | Cuối tuần | 69.900 | 74.046 | 76.200 | — | — | 73.382 | 51.920 | 64.760 | -29.2% | -11.7% |
| 56 | Bike | 15 | Lễ Tết | 69.900 | 74.046 | 76.200 | — | — | 84.389 | 59.408 | 74.324 | -29.6% | -11.9% |
| 57 | Bike | 20 | Nội thành | 91.400 | 97.211 | 100.200 | — | — | 96.270 | 66.560 | 85.180 | -30.9% | -11.5% |
| 58 | Bike | 20 | Ngoại thành | 91.400 | 97.211 | 100.200 | — | — | 96.270 | 66.560 | 85.180 | -30.9% | -11.5% |
| 59 | Bike | 20 | Sân bay | 91.400 | 97.211 | 100.200 | — | — | 96.270 | 76.560 | 95.180 | -20.5% | -1.1% |
| 60 | Bike | 20 | Giờ cao điểm | 91.400 | 97.211 | 100.200 | — | — | 110.711 | 73.016 | 93.598 | -34.0% | -15.5% |
| 61 | Bike | 20 | Mưa | 91.400 | 97.211 | 100.200 | — | — | 110.711 | 76.244 | 97.807 | -31.1% | -11.7% |
| 62 | Bike | 20 | Đêm | 91.400 | 97.211 | 100.200 | — | — | 110.711 | 79.472 | 102.016 | -28.2% | -7.9% |
| 63 | Bike | 20 | Cuối tuần | 91.400 | 97.211 | 100.200 | — | — | 96.270 | 66.560 | 85.180 | -30.9% | -11.5% |
| 64 | Bike | 20 | Lễ Tết | 91.400 | 97.211 | 100.200 | — | — | 110.711 | 76.244 | 97.807 | -31.1% | -11.7% |
| 65 | Bike | 25 | Nội thành | 112.900 | 120.376 | 124.200 | — | — | 119.159 | 81.200 | 105.600 | -31.9% | -11.4% |
| 66 | Bike | 25 | Ngoại thành | 112.900 | 120.376 | 124.200 | — | — | 119.159 | 81.200 | 105.600 | -31.9% | -11.4% |
| 67 | Bike | 25 | Sân bay | 112.900 | 120.376 | 124.200 | — | — | 119.159 | 91.200 | 115.600 | -23.5% | -3.0% |
| 68 | Bike | 25 | Giờ cao điểm | 112.900 | 120.376 | 124.200 | — | — | 137.032 | 89.120 | 116.060 | -35.0% | -15.3% |
| 69 | Bike | 25 | Mưa | 112.900 | 120.376 | 124.200 | — | — | 137.032 | 93.080 | 121.290 | -32.1% | -11.5% |
| 70 | Bike | 25 | Đêm | 112.900 | 120.376 | 124.200 | — | — | 137.032 | 97.040 | 126.520 | -29.2% | -7.7% |
| 71 | Bike | 25 | Cuối tuần | 112.900 | 120.376 | 124.200 | — | — | 119.159 | 81.200 | 105.600 | -31.9% | -11.4% |
| 72 | Bike | 25 | Lễ Tết | 112.900 | 120.376 | 124.200 | — | — | 137.032 | 93.080 | 121.290 | -32.1% | -11.5% |
| 73 | Bike | 30 | Nội thành | 134.400 | 143.541 | 148.200 | — | — | 142.047 | 95.840 | 126.020 | -32.5% | -11.3% |
| 74 | Bike | 30 | Ngoại thành | 134.400 | 143.541 | 148.200 | — | — | 142.047 | 95.840 | 126.020 | -32.5% | -11.3% |
| 75 | Bike | 30 | Sân bay | 134.400 | 143.541 | 148.200 | — | — | 142.047 | 105.840 | 136.020 | -25.5% | -4.2% |
| 76 | Bike | 30 | Giờ cao điểm | 134.400 | 143.541 | 148.200 | — | — | 163.354 | 105.224 | 138.522 | -35.6% | -15.2% |
| 77 | Bike | 30 | Mưa | 134.400 | 143.541 | 148.200 | — | — | 163.354 | 109.916 | 144.773 | -32.7% | -11.4% |
| 78 | Bike | 30 | Đêm | 134.400 | 143.541 | 148.200 | — | — | 163.354 | 114.608 | 151.024 | -29.8% | -7.5% |
| 79 | Bike | 30 | Cuối tuần | 134.400 | 143.541 | 148.200 | — | — | 142.047 | 95.840 | 126.020 | -32.5% | -11.3% |
| 80 | Bike | 30 | Lễ Tết | 134.400 | 143.541 | 148.200 | — | — | 163.354 | 109.916 | 144.773 | -32.7% | -11.4% |
| 81 | Bike | 40 | Nội thành | 177.400 | 189.871 | 196.200 | — | — | 187.824 | 125.120 | 166.860 | -33.4% | -11.2% |
| 82 | Bike | 40 | Ngoại thành | 177.400 | 189.871 | 196.200 | — | — | 187.824 | 125.120 | 166.860 | -33.4% | -11.2% |
| 83 | Bike | 40 | Sân bay | 177.400 | 189.871 | 196.200 | — | — | 187.824 | 135.120 | 176.860 | -28.1% | -5.8% |
| 84 | Bike | 40 | Giờ cao điểm | 177.400 | 189.871 | 196.200 | — | — | 215.997 | 137.432 | 183.446 | -36.4% | -15.1% |
| 85 | Bike | 40 | Mưa | 177.400 | 189.871 | 196.200 | — | — | 215.997 | 143.588 | 191.739 | -33.5% | -11.2% |
| 86 | Bike | 40 | Đêm | 177.400 | 189.871 | 196.200 | — | — | 215.997 | 149.744 | 200.032 | -30.7% | -7.4% |
| 87 | Bike | 40 | Cuối tuần | 177.400 | 189.871 | 196.200 | — | — | 187.824 | 125.120 | 166.860 | -33.4% | -11.2% |
| 88 | Bike | 40 | Lễ Tết | 177.400 | 189.871 | 196.200 | — | — | 215.997 | 143.588 | 191.739 | -33.5% | -11.2% |
| 89 | Bike | 50 | Nội thành | 220.400 | 236.201 | 244.200 | — | — | 233.600 | 154.400 | 207.700 | -33.9% | -11.1% |
| 90 | Bike | 50 | Ngoại thành | 220.400 | 236.201 | 244.200 | — | — | 233.600 | 154.400 | 207.700 | -33.9% | -11.1% |
| 91 | Bike | 50 | Sân bay | 220.400 | 236.201 | 244.200 | — | — | 233.600 | 164.400 | 217.700 | -29.6% | -6.8% |
| 92 | Bike | 50 | Giờ cao điểm | 220.400 | 236.201 | 244.200 | — | — | 268.640 | 169.640 | 228.370 | -36.9% | -15.0% |
| 93 | Bike | 50 | Mưa | 220.400 | 236.201 | 244.200 | — | — | 268.640 | 177.260 | 238.705 | -34.0% | -11.1% |
| 94 | Bike | 50 | Đêm | 220.400 | 236.201 | 244.200 | — | — | 268.640 | 184.880 | 249.040 | -31.2% | -7.3% |
| 95 | Bike | 50 | Cuối tuần | 220.400 | 236.201 | 244.200 | — | — | 233.600 | 154.400 | 207.700 | -33.9% | -11.1% |
| 96 | Bike | 50 | Lễ Tết | 220.400 | 236.201 | 244.200 | — | — | 268.640 | 177.260 | 238.705 | -34.0% | -11.1% |
| 97 | Bike | 60 | Nội thành | 263.400 | 282.531 | 292.200 | — | — | 279.377 | 183.680 | 248.540 | -34.3% | -11.0% |
| 98 | Bike | 60 | Ngoại thành | 263.400 | 282.531 | 292.200 | — | — | 279.377 | 183.680 | 248.540 | -34.3% | -11.0% |
| 99 | Bike | 60 | Sân bay | 263.400 | 282.531 | 292.200 | — | — | 279.377 | 193.680 | 258.540 | -30.7% | -7.5% |
| 100 | Bike | 60 | Giờ cao điểm | 263.400 | 282.531 | 292.200 | — | — | 321.284 | 201.848 | 273.294 | -37.2% | -14.9% |
| 101 | Bike | 60 | Mưa | 263.400 | 282.531 | 292.200 | — | — | 321.284 | 210.932 | 285.671 | -34.3% | -11.1% |
| 102 | Bike | 60 | Đêm | 263.400 | 282.531 | 292.200 | — | — | 321.284 | 220.016 | 298.048 | -31.5% | -7.2% |
| 103 | Bike | 60 | Cuối tuần | 263.400 | 282.531 | 292.200 | — | — | 279.377 | 183.680 | 248.540 | -34.3% | -11.0% |
| 104 | Bike | 60 | Lễ Tết | 263.400 | 282.531 | 292.200 | — | — | 321.284 | 210.932 | 285.671 | -34.3% | -11.1% |
| 105 | Car | 2 | Nội thành | 29.000 | 43.306 | 35.000 | 31.600 | 32.750 | 35.769 | 27.000 | 35.576 | -24.5% | -0.5% |
| 106 | Car | 2 | Ngoại thành | 29.000 | 43.306 | 35.000 | 31.600 | 32.750 | 35.769 | 27.000 | 35.576 | -24.5% | -0.5% |
| 107 | Car | 2 | Sân bay | 29.000 | 43.306 | 35.000 | 37.600 | 32.750 | 50.769 | 37.000 | 45.576 | -27.1% | -10.2% |
| 108 | Car | 2 | Giờ cao điểm | 29.000 | 43.306 | 35.000 | 34.760 | 32.750 | 41.134 | 27.000 | 38.834 | -34.4% | -5.6% |
| 109 | Car | 2 | Mưa | 29.000 | 43.306 | 35.000 | 31.600 | 32.750 | 41.134 | 27.000 | 40.462 | -34.4% | -1.6% |
| 110 | Car | 2 | Đêm | 29.000 | 43.306 | 35.000 | 31.600 | 32.750 | 41.134 | 27.000 | 42.091 | -34.4% | 2.3% |
| 111 | Car | 2 | Cuối tuần | 29.000 | 43.306 | 35.000 | 31.600 | 32.750 | 35.769 | 27.000 | 35.576 | -24.5% | -0.5% |
| 112 | Car | 2 | Lễ Tết | 29.000 | 43.306 | 35.000 | 31.600 | 32.750 | 41.134 | 27.000 | 40.462 | -34.4% | -1.6% |
| 113 | Car | 4 | Nội thành | 49.000 | 66.916 | 65.000 | 59.600 | 61.750 | 60.305 | 31.520 | 55.152 | -47.7% | -8.5% |
| 114 | Car | 4 | Ngoại thành | 49.000 | 66.916 | 65.000 | 59.600 | 61.750 | 60.305 | 31.520 | 55.152 | -47.7% | -8.5% |
| 115 | Car | 4 | Sân bay | 49.000 | 66.916 | 65.000 | 71.600 | 61.750 | 75.305 | 41.520 | 65.152 | -44.9% | -13.5% |
| 116 | Car | 4 | Giờ cao điểm | 49.000 | 66.916 | 65.000 | 65.560 | 61.750 | 69.351 | 34.472 | 60.367 | -50.3% | -13.0% |
| 117 | Car | 4 | Mưa | 49.000 | 66.916 | 65.000 | 59.600 | 61.750 | 69.351 | 35.948 | 62.975 | -48.2% | -9.2% |
| 118 | Car | 4 | Đêm | 49.000 | 66.916 | 65.000 | 59.600 | 61.750 | 69.351 | 37.424 | 65.582 | -46.0% | -5.4% |
| 119 | Car | 4 | Cuối tuần | 49.000 | 66.916 | 65.000 | 59.600 | 61.750 | 60.305 | 31.520 | 55.152 | -47.7% | -8.5% |
| 120 | Car | 4 | Lễ Tết | 49.000 | 66.916 | 65.000 | 59.600 | 61.750 | 69.351 | 35.948 | 62.975 | -48.2% | -9.2% |
| 121 | Car | 6 | Nội thành | 69.000 | 90.526 | 95.000 | 87.600 | 90.750 | 84.842 | 41.280 | 74.728 | -51.3% | -11.9% |
| 122 | Car | 6 | Ngoại thành | 69.000 | 90.526 | 95.000 | 87.600 | 90.750 | 84.842 | 41.280 | 74.728 | -51.3% | -11.9% |
| 123 | Car | 6 | Sân bay | 69.000 | 90.526 | 95.000 | 105.600 | 90.750 | 99.842 | 51.280 | 84.728 | -48.6% | -15.1% |
| 124 | Car | 6 | Giờ cao điểm | 69.000 | 90.526 | 95.000 | 96.360 | 90.750 | 97.568 | 45.208 | 81.901 | -53.7% | -16.1% |
| 125 | Car | 6 | Mưa | 69.000 | 90.526 | 95.000 | 87.600 | 90.750 | 97.568 | 47.172 | 85.487 | -51.7% | -12.4% |
| 126 | Car | 6 | Đêm | 69.000 | 90.526 | 95.000 | 87.600 | 90.750 | 97.568 | 49.136 | 89.074 | -49.6% | -8.7% |
| 127 | Car | 6 | Cuối tuần | 69.000 | 90.526 | 95.000 | 87.600 | 90.750 | 84.842 | 41.280 | 74.728 | -51.3% | -11.9% |
| 128 | Car | 6 | Lễ Tết | 69.000 | 90.526 | 95.000 | 87.600 | 90.750 | 97.568 | 47.172 | 85.487 | -51.7% | -12.4% |
| 129 | Car | 8 | Nội thành | 89.000 | 114.136 | 125.000 | 115.600 | 119.750 | 109.379 | 51.040 | 94.304 | -53.3% | -13.8% |
| 130 | Car | 8 | Ngoại thành | 89.000 | 114.136 | 125.000 | 115.600 | 119.750 | 109.379 | 51.040 | 94.304 | -53.3% | -13.8% |
| 131 | Car | 8 | Sân bay | 89.000 | 114.136 | 125.000 | 139.600 | 119.750 | 124.379 | 61.040 | 104.304 | -50.9% | -16.1% |
| 132 | Car | 8 | Giờ cao điểm | 89.000 | 114.136 | 125.000 | 127.160 | 119.750 | 125.785 | 55.944 | 103.434 | -55.5% | -17.8% |
| 133 | Car | 8 | Mưa | 89.000 | 114.136 | 125.000 | 115.600 | 119.750 | 125.785 | 58.396 | 108.000 | -53.6% | -14.1% |
| 134 | Car | 8 | Đêm | 89.000 | 114.136 | 125.000 | 115.600 | 119.750 | 125.785 | 60.848 | 112.565 | -51.6% | -10.5% |
| 135 | Car | 8 | Cuối tuần | 89.000 | 114.136 | 125.000 | 115.600 | 119.750 | 109.379 | 51.040 | 94.304 | -53.3% | -13.8% |
| 136 | Car | 8 | Lễ Tết | 89.000 | 114.136 | 125.000 | 115.600 | 119.750 | 125.785 | 58.396 | 108.000 | -53.6% | -14.1% |
| 137 | Car | 10 | Nội thành | 109.000 | 137.746 | 155.000 | 143.600 | 148.750 | 133.915 | 60.800 | 113.880 | -54.6% | -15.0% |
| 138 | Car | 10 | Ngoại thành | 109.000 | 137.746 | 155.000 | 143.600 | 148.750 | 133.915 | 60.800 | 113.880 | -54.6% | -15.0% |
| 139 | Car | 10 | Sân bay | 109.000 | 137.746 | 155.000 | 173.600 | 148.750 | 148.915 | 70.800 | 123.880 | -52.5% | -16.8% |
| 140 | Car | 10 | Giờ cao điểm | 109.000 | 137.746 | 155.000 | 157.960 | 148.750 | 154.003 | 66.680 | 124.968 | -56.7% | -18.9% |
| 141 | Car | 10 | Mưa | 109.000 | 137.746 | 155.000 | 143.600 | 148.750 | 154.003 | 69.620 | 130.512 | -54.8% | -15.3% |
| 142 | Car | 10 | Đêm | 109.000 | 137.746 | 155.000 | 143.600 | 148.750 | 154.003 | 72.560 | 136.056 | -52.9% | -11.7% |
| 143 | Car | 10 | Cuối tuần | 109.000 | 137.746 | 155.000 | 143.600 | 148.750 | 133.915 | 60.800 | 113.880 | -54.6% | -15.0% |
| 144 | Car | 10 | Lễ Tết | 109.000 | 137.746 | 155.000 | 143.600 | 148.750 | 154.003 | 69.620 | 130.512 | -54.8% | -15.3% |
| 145 | Car | 12 | Nội thành | 129.000 | 159.304 | 185.000 | 171.600 | 177.750 | 157.768 | 70.560 | 133.456 | -55.3% | -15.4% |
| 146 | Car | 12 | Ngoại thành | 129.000 | 159.304 | 185.000 | 171.600 | 177.750 | 157.768 | 70.560 | 133.456 | -55.3% | -15.4% |
| 147 | Car | 12 | Sân bay | 129.000 | 159.304 | 185.000 | 207.600 | 177.750 | 172.768 | 80.560 | 143.456 | -53.4% | -17.0% |
| 148 | Car | 12 | Giờ cao điểm | 129.000 | 159.304 | 185.000 | 188.760 | 177.750 | 181.433 | 77.416 | 146.502 | -57.3% | -19.3% |
| 149 | Car | 12 | Mưa | 129.000 | 159.304 | 185.000 | 171.600 | 177.750 | 181.433 | 80.844 | 153.024 | -55.4% | -15.7% |
| 150 | Car | 12 | Đêm | 129.000 | 159.304 | 185.000 | 171.600 | 177.750 | 181.433 | 84.272 | 159.547 | -53.6% | -12.1% |
| 151 | Car | 12 | Cuối tuần | 129.000 | 159.304 | 185.000 | 171.600 | 177.750 | 157.768 | 70.560 | 133.456 | -55.3% | -15.4% |
| 152 | Car | 12 | Lễ Tết | 129.000 | 159.304 | 185.000 | 171.600 | 177.750 | 181.433 | 80.844 | 153.024 | -55.4% | -15.7% |
| 153 | Car | 15 | Nội thành | 159.000 | 191.641 | 230.000 | 213.600 | 221.250 | 193.547 | 85.200 | 162.820 | -56.0% | -15.9% |
| 154 | Car | 15 | Ngoại thành | 159.000 | 191.641 | 230.000 | 213.600 | 221.250 | 193.547 | 85.200 | 162.820 | -56.0% | -15.9% |
| 155 | Car | 15 | Sân bay | 159.000 | 191.641 | 230.000 | 258.600 | 221.250 | 208.547 | 95.200 | 172.820 | -54.4% | -17.1% |
| 156 | Car | 15 | Giờ cao điểm | 159.000 | 191.641 | 230.000 | 234.960 | 221.250 | 222.579 | 93.520 | 178.802 | -58.0% | -19.7% |
| 157 | Car | 15 | Mưa | 159.000 | 191.641 | 230.000 | 213.600 | 221.250 | 222.579 | 97.680 | 186.793 | -56.1% | -16.1% |
| 158 | Car | 15 | Đêm | 159.000 | 191.641 | 230.000 | 213.600 | 221.250 | 222.579 | 101.840 | 194.784 | -54.2% | -12.5% |
| 159 | Car | 15 | Cuối tuần | 159.000 | 191.641 | 230.000 | 213.600 | 221.250 | 193.547 | 85.200 | 162.820 | -56.0% | -15.9% |
| 160 | Car | 15 | Lễ Tết | 159.000 | 191.641 | 230.000 | 213.600 | 221.250 | 222.579 | 97.680 | 186.793 | -56.1% | -16.1% |
| 161 | Car | 20 | Nội thành | 209.000 | 245.536 | 305.000 | 283.600 | 293.750 | 253.179 | 109.600 | 211.760 | -56.7% | -16.4% |
| 162 | Car | 20 | Ngoại thành | 209.000 | 245.536 | 305.000 | 283.600 | 293.750 | 253.179 | 109.600 | 211.760 | -56.7% | -16.4% |
| 163 | Car | 20 | Sân bay | 209.000 | 245.536 | 305.000 | 343.600 | 293.750 | 268.179 | 119.600 | 221.760 | -55.4% | -17.3% |
| 164 | Car | 20 | Giờ cao điểm | 209.000 | 245.536 | 305.000 | 311.960 | 293.750 | 291.155 | 120.360 | 232.636 | -58.7% | -20.1% |
| 165 | Car | 20 | Mưa | 209.000 | 245.536 | 305.000 | 283.600 | 293.750 | 291.155 | 125.740 | 243.074 | -56.8% | -16.5% |
| 166 | Car | 20 | Đêm | 209.000 | 245.536 | 305.000 | 283.600 | 293.750 | 291.155 | 131.120 | 253.512 | -55.0% | -12.9% |
| 167 | Car | 20 | Cuối tuần | 209.000 | 245.536 | 305.000 | 283.600 | 293.750 | 253.179 | 109.600 | 211.760 | -56.7% | -16.4% |
| 168 | Car | 20 | Lễ Tết | 209.000 | 245.536 | 305.000 | 283.600 | 293.750 | 291.155 | 125.740 | 243.074 | -56.8% | -16.5% |
| 169 | Car | 25 | Nội thành | 259.000 | 299.431 | 377.000 | 353.600 | 366.250 | 311.810 | 134.000 | 260.700 | -57.0% | -16.4% |
| 170 | Car | 25 | Ngoại thành | 259.000 | 299.431 | 377.000 | 353.600 | 366.250 | 311.810 | 134.000 | 260.700 | -57.0% | -16.4% |
| 171 | Car | 25 | Sân bay | 259.000 | 299.431 | 377.000 | 428.600 | 366.250 | 326.810 | 144.000 | 270.700 | -55.9% | -17.2% |
| 172 | Car | 25 | Giờ cao điểm | 259.000 | 299.431 | 377.000 | 388.960 | 366.250 | 358.582 | 147.200 | 286.470 | -58.9% | -20.1% |
| 173 | Car | 25 | Mưa | 259.000 | 299.431 | 377.000 | 353.600 | 366.250 | 358.582 | 153.800 | 299.355 | -57.1% | -16.5% |
| 174 | Car | 25 | Đêm | 259.000 | 299.431 | 377.000 | 353.600 | 366.250 | 358.582 | 160.400 | 312.240 | -55.3% | -12.9% |
| 175 | Car | 25 | Cuối tuần | 259.000 | 299.431 | 377.000 | 353.600 | 366.250 | 311.810 | 134.000 | 260.700 | -57.0% | -16.4% |
| 176 | Car | 25 | Lễ Tết | 259.000 | 299.431 | 377.000 | 353.600 | 366.250 | 358.582 | 153.800 | 299.355 | -57.1% | -16.5% |
| 177 | Car | 30 | Nội thành | 309.000 | 353.326 | 437.000 | 423.600 | 438.750 | 366.442 | 158.400 | 309.640 | -56.8% | -15.5% |
| 178 | Car | 30 | Ngoại thành | 309.000 | 353.326 | 437.000 | 423.600 | 438.750 | 366.442 | 158.400 | 309.640 | -56.8% | -15.5% |
| 179 | Car | 30 | Sân bay | 309.000 | 353.326 | 437.000 | 513.600 | 438.750 | 381.442 | 168.400 | 319.640 | -55.9% | -16.2% |
| 180 | Car | 30 | Giờ cao điểm | 309.000 | 353.326 | 437.000 | 465.960 | 438.750 | 421.408 | 174.040 | 340.304 | -58.7% | -19.2% |
| 181 | Car | 30 | Mưa | 309.000 | 353.326 | 437.000 | 423.600 | 438.750 | 421.408 | 181.860 | 355.636 | -56.8% | -15.6% |
| 182 | Car | 30 | Đêm | 309.000 | 353.326 | 437.000 | 423.600 | 438.750 | 421.408 | 189.680 | 370.968 | -55.0% | -12.0% |
| 183 | Car | 30 | Cuối tuần | 309.000 | 353.326 | 437.000 | 423.600 | 438.750 | 366.442 | 158.400 | 309.640 | -56.8% | -15.5% |
| 184 | Car | 30 | Lễ Tết | 309.000 | 353.326 | 437.000 | 423.600 | 438.750 | 421.408 | 181.860 | 355.636 | -56.8% | -15.6% |
| 185 | Car | 40 | Nội thành | 409.000 | 461.116 | 557.000 | 533.600 | 557.650 | 475.705 | 207.200 | 407.520 | -56.4% | -14.3% |
| 186 | Car | 40 | Ngoại thành | 409.000 | 461.116 | 557.000 | 533.600 | 557.650 | 475.705 | 207.200 | 407.520 | -56.4% | -14.3% |
| 187 | Car | 40 | Sân bay | 409.000 | 461.116 | 557.000 | 653.600 | 557.650 | 490.705 | 217.200 | 417.520 | -55.7% | -14.9% |
| 188 | Car | 40 | Giờ cao điểm | 409.000 | 461.116 | 557.000 | 586.960 | 557.650 | 547.061 | 227.720 | 447.972 | -58.4% | -18.1% |
| 189 | Car | 40 | Mưa | 409.000 | 461.116 | 557.000 | 533.600 | 557.650 | 547.061 | 237.980 | 468.198 | -56.5% | -14.4% |
| 190 | Car | 40 | Đêm | 409.000 | 461.116 | 557.000 | 533.600 | 557.650 | 547.061 | 248.240 | 488.424 | -54.6% | -10.7% |
| 191 | Car | 40 | Cuối tuần | 409.000 | 461.116 | 557.000 | 533.600 | 557.650 | 475.705 | 207.200 | 407.520 | -56.4% | -14.3% |
| 192 | Car | 40 | Lễ Tết | 409.000 | 461.116 | 557.000 | 533.600 | 557.650 | 547.061 | 237.980 | 468.198 | -56.5% | -14.4% |
| 193 | Car | 50 | Nội thành | 509.000 | 568.906 | 677.000 | 643.600 | 673.650 | 584.969 | 256.000 | **505.400** | -56.2% | -13.6% |
| 194 | Car | 50 | Ngoại thành | 509.000 | 568.906 | 677.000 | 643.600 | 673.650 | 584.969 | 256.000 | **505.400** | -56.2% | -13.6% |
| 195 | Car | 50 | Sân bay | 509.000 | 568.906 | 677.000 | 793.600 | 673.650 | 599.969 | 266.000 | **515.400** | -55.7% | -14.1% |
| 196 | Car | 50 | Giờ cao điểm | 509.000 | 568.906 | 677.000 | 707.960 | 673.650 | 672.714 | 281.400 | **555.640** | -58.2% | -17.4% |
| 197 | Car | 50 | Mưa | 509.000 | 568.906 | 677.000 | 643.600 | 673.650 | 672.714 | 294.100 | **580.760** | -56.3% | -13.7% |
| 198 | Car | 50 | Đêm | 509.000 | 568.906 | 677.000 | 643.600 | 673.650 | 672.714 | 306.800 | **605.880** | -54.4% | -9.9% |
| 199 | Car | 50 | Cuối tuần | 509.000 | 568.906 | 677.000 | 643.600 | 673.650 | 584.969 | 256.000 | **505.400** | -56.2% | -13.6% |
| 200 | Car | 50 | Lễ Tết | 509.000 | 568.906 | 677.000 | 643.600 | 673.650 | 672.714 | 294.100 | **580.760** | -56.3% | -13.7% |
| 201 | Car | 60 | Nội thành | 609.000 | 676.696 | 797.000 | 753.600 | 789.650 | 694.232 | 304.800 | **603.280** | -56.1% | -13.1% |
| 202 | Car | 60 | Ngoại thành | 609.000 | 676.696 | 797.000 | 753.600 | 789.650 | 694.232 | 304.800 | **603.280** | -56.1% | -13.1% |
| 203 | Car | 60 | Sân bay | 609.000 | 676.696 | 797.000 | 933.600 | 789.650 | 709.232 | 314.800 | **613.280** | -55.6% | -13.5% |
| 204 | Car | 60 | Giờ cao điểm | 609.000 | 676.696 | 797.000 | 828.960 | 789.650 | 798.367 | 335.080 | **663.308** | -58.0% | -16.9% |
| 205 | Car | 60 | Mưa | 609.000 | 676.696 | 797.000 | 753.600 | 789.650 | 798.367 | 350.220 | **693.322** | -56.1% | -13.2% |
| 206 | Car | 60 | Đêm | 609.000 | 676.696 | 797.000 | 753.600 | 789.650 | 798.367 | 365.360 | **723.336** | -54.2% | -9.4% |
| 207 | Car | 60 | Cuối tuần | 609.000 | 676.696 | 797.000 | 753.600 | 789.650 | 694.232 | 304.800 | **603.280** | -56.1% | -13.1% |
| 208 | Car | 60 | Lễ Tết | 609.000 | 676.696 | 797.000 | 753.600 | 789.650 | 798.367 | 350.220 | **693.322** | -56.1% | -13.2% |
| 209 | XL | 2 | Nội thành | 37.700 | 56.298 | 50.000 | 41.080 | 42.575 | 47.999 | 42.000 | 51.080 | -12.5% | 6.4% |
| 210 | XL | 2 | Ngoại thành | 37.700 | 56.298 | 50.000 | 41.080 | 42.575 | 47.999 | 42.000 | 51.080 | -12.5% | 6.4% |
| 211 | XL | 2 | Sân bay | 37.700 | 56.298 | 50.000 | 48.880 | 42.575 | 47.999 | 52.000 | 61.080 | 8.3% | 27.3% |
| 212 | XL | 2 | Giờ cao điểm | 37.700 | 56.298 | 50.000 | 45.188 | 42.575 | 55.199 | 42.000 | 55.888 | -23.9% | 1.2% |
| 213 | XL | 2 | Mưa | 37.700 | 56.298 | 50.000 | 41.080 | 42.575 | 55.199 | 42.000 | 58.292 | -23.9% | 5.6% |
| 214 | XL | 2 | Đêm | 37.700 | 56.298 | 50.000 | 41.080 | 42.575 | 55.199 | 42.000 | 60.696 | -23.9% | 10.0% |
| 215 | XL | 2 | Cuối tuần | 37.700 | 56.298 | 50.000 | 41.080 | 42.575 | 47.999 | 42.000 | 51.080 | -12.5% | 6.4% |
| 216 | XL | 2 | Lễ Tết | 37.700 | 56.298 | 50.000 | 41.080 | 42.575 | 55.199 | 42.000 | 58.292 | -23.9% | 5.6% |
| 217 | XL | 4 | Nội thành | 63.700 | 86.991 | 84.000 | 77.480 | 80.275 | 78.230 | 44.400 | 77.160 | -43.2% | -1.4% |
| 218 | XL | 4 | Ngoại thành | 63.700 | 86.991 | 84.000 | 77.480 | 80.275 | 78.230 | 44.400 | 77.160 | -43.2% | -1.4% |
| 219 | XL | 4 | Sân bay | 63.700 | 86.991 | 84.000 | 93.080 | 80.275 | 78.230 | 54.400 | 87.160 | -30.5% | 11.4% |
| 220 | XL | 4 | Giờ cao điểm | 63.700 | 86.991 | 84.000 | 85.228 | 80.275 | 89.965 | 48.640 | 84.576 | -45.9% | -6.0% |
| 221 | XL | 4 | Mưa | 63.700 | 86.991 | 84.000 | 77.480 | 80.275 | 89.965 | 50.760 | 88.284 | -43.6% | -1.9% |
| 222 | XL | 4 | Đêm | 63.700 | 86.991 | 84.000 | 77.480 | 80.275 | 89.965 | 52.880 | 91.992 | -41.2% | 2.3% |
| 223 | XL | 4 | Cuối tuần | 63.700 | 86.991 | 84.000 | 77.480 | 80.275 | 78.230 | 44.400 | 77.160 | -43.2% | -1.4% |
| 224 | XL | 4 | Lễ Tết | 63.700 | 86.991 | 84.000 | 77.480 | 80.275 | 89.965 | 50.760 | 88.284 | -43.6% | -1.9% |
| 225 | XL | 6 | Nội thành | 89.700 | 117.684 | 126.000 | 113.880 | 117.975 | 111.128 | 56.600 | 103.240 | -49.1% | -7.1% |
| 226 | XL | 6 | Ngoại thành | 89.700 | 117.684 | 126.000 | 113.880 | 117.975 | 111.128 | 56.600 | 103.240 | -49.1% | -7.1% |
| 227 | XL | 6 | Sân bay | 89.700 | 117.684 | 126.000 | 137.280 | 117.975 | 111.128 | 66.600 | 113.240 | -40.1% | 1.9% |
| 228 | XL | 6 | Giờ cao điểm | 89.700 | 117.684 | 126.000 | 125.268 | 117.975 | 127.797 | 62.060 | 113.264 | -51.4% | -11.4% |
| 229 | XL | 6 | Mưa | 89.700 | 117.684 | 126.000 | 113.880 | 117.975 | 127.797 | 64.790 | 118.276 | -49.3% | -7.5% |
| 230 | XL | 6 | Đêm | 89.700 | 117.684 | 126.000 | 113.880 | 117.975 | 127.797 | 67.520 | 123.288 | -47.2% | -3.5% |
| 231 | XL | 6 | Cuối tuần | 89.700 | 117.684 | 126.000 | 113.880 | 117.975 | 111.128 | 56.600 | 103.240 | -49.1% | -7.1% |
| 232 | XL | 6 | Lễ Tết | 89.700 | 117.684 | 126.000 | 113.880 | 117.975 | 127.797 | 64.790 | 118.276 | -49.3% | -7.5% |
| 233 | XL | 8 | Nội thành | 115.700 | 148.377 | 168.000 | 150.280 | 155.675 | 144.026 | 68.800 | 129.320 | -52.2% | -10.2% |
| 234 | XL | 8 | Ngoại thành | 115.700 | 148.377 | 168.000 | 150.280 | 155.675 | 144.026 | 68.800 | 129.320 | -52.2% | -10.2% |
| 235 | XL | 8 | Sân bay | 115.700 | 148.377 | 168.000 | 181.480 | 155.675 | 144.026 | 78.800 | 139.320 | -45.3% | -3.3% |
| 236 | XL | 8 | Giờ cao điểm | 115.700 | 148.377 | 168.000 | 165.308 | 155.675 | 165.629 | 75.480 | 141.952 | -54.4% | -14.3% |
| 237 | XL | 8 | Mưa | 115.700 | 148.377 | 168.000 | 150.280 | 155.675 | 165.629 | 78.820 | 148.268 | -52.4% | -10.5% |
| 238 | XL | 8 | Đêm | 115.700 | 148.377 | 168.000 | 150.280 | 155.675 | 165.629 | 82.160 | 154.584 | -50.4% | -6.7% |
| 239 | XL | 8 | Cuối tuần | 115.700 | 148.377 | 168.000 | 150.280 | 155.675 | 144.026 | 68.800 | 129.320 | -52.2% | -10.2% |
| 240 | XL | 8 | Lễ Tết | 115.700 | 148.377 | 168.000 | 150.280 | 155.675 | 165.629 | 78.820 | 148.268 | -52.4% | -10.5% |
| 241 | XL | 10 | Nội thành | 141.700 | 179.070 | 210.000 | 186.680 | 193.375 | 176.923 | 81.000 | 155.400 | -54.2% | -12.2% |
| 242 | XL | 10 | Ngoại thành | 141.700 | 179.070 | 210.000 | 186.680 | 193.375 | 176.923 | 81.000 | 155.400 | -54.2% | -12.2% |
| 243 | XL | 10 | Sân bay | 141.700 | 179.070 | 210.000 | 225.680 | 193.375 | 176.923 | 91.000 | 165.400 | -48.6% | -6.5% |
| 244 | XL | 10 | Giờ cao điểm | 141.700 | 179.070 | 210.000 | 205.348 | 193.375 | 203.462 | 88.900 | 170.640 | -56.3% | -16.1% |
| 245 | XL | 10 | Mưa | 141.700 | 179.070 | 210.000 | 186.680 | 193.375 | 203.462 | 92.850 | 178.260 | -54.4% | -12.4% |
| 246 | XL | 10 | Đêm | 141.700 | 179.070 | 210.000 | 186.680 | 193.375 | 203.462 | 96.800 | 185.880 | -52.4% | -8.6% |
| 247 | XL | 10 | Cuối tuần | 141.700 | 179.070 | 210.000 | 186.680 | 193.375 | 176.923 | 81.000 | 155.400 | -54.2% | -12.2% |
| 248 | XL | 10 | Lễ Tết | 141.700 | 179.070 | 210.000 | 186.680 | 193.375 | 203.462 | 92.850 | 178.260 | -54.4% | -12.4% |
| 249 | XL | 12 | Nội thành | 167.700 | 207.095 | 252.000 | 223.080 | 231.075 | 208.932 | 93.200 | 181.480 | -55.4% | -13.1% |
| 250 | XL | 12 | Ngoại thành | 167.700 | 207.095 | 252.000 | 223.080 | 231.075 | 208.932 | 93.200 | 181.480 | -55.4% | -13.1% |
| 251 | XL | 12 | Sân bay | 167.700 | 207.095 | 252.000 | 269.880 | 231.075 | 208.932 | 103.200 | 191.480 | -50.6% | -8.4% |
| 252 | XL | 12 | Giờ cao điểm | 167.700 | 207.095 | 252.000 | 245.388 | 231.075 | 240.271 | 102.320 | 199.328 | -57.4% | -17.0% |
| 253 | XL | 12 | Mưa | 167.700 | 207.095 | 252.000 | 223.080 | 231.075 | 240.271 | 106.880 | 208.252 | -55.5% | -13.3% |
| 254 | XL | 12 | Đêm | 167.700 | 207.095 | 252.000 | 223.080 | 231.075 | 240.271 | 111.440 | 217.176 | -53.6% | -9.6% |
| 255 | XL | 12 | Cuối tuần | 167.700 | 207.095 | 252.000 | 223.080 | 231.075 | 208.932 | 93.200 | 181.480 | -55.4% | -13.1% |
| 256 | XL | 12 | Lễ Tết | 167.700 | 207.095 | 252.000 | 223.080 | 231.075 | 240.271 | 106.880 | 208.252 | -55.5% | -13.3% |
| 257 | XL | 15 | Nội thành | 206.700 | 249.133 | 315.000 | 277.680 | 287.625 | 256.944 | 111.500 | 220.600 | -56.6% | -14.1% |
| 258 | XL | 15 | Ngoại thành | 206.700 | 249.133 | 315.000 | 277.680 | 287.625 | 256.944 | 111.500 | 220.600 | -56.6% | -14.1% |
| 259 | XL | 15 | Sân bay | 206.700 | 249.133 | 315.000 | 336.180 | 287.625 | 256.944 | 121.500 | 230.600 | -52.7% | -10.3% |
| 260 | XL | 15 | Giờ cao điểm | 206.700 | 249.133 | 315.000 | 305.448 | 287.625 | 295.486 | 122.450 | 242.360 | -58.6% | -18.0% |
| 261 | XL | 15 | Mưa | 206.700 | 249.133 | 315.000 | 277.680 | 287.625 | 295.486 | 127.925 | 253.240 | -56.7% | -14.3% |
| 262 | XL | 15 | Đêm | 206.700 | 249.133 | 315.000 | 277.680 | 287.625 | 295.486 | 133.400 | 264.120 | -54.9% | -10.6% |
| 263 | XL | 15 | Cuối tuần | 206.700 | 249.133 | 315.000 | 277.680 | 287.625 | 256.944 | 111.500 | 220.600 | -56.6% | -14.1% |
| 264 | XL | 15 | Lễ Tết | 206.700 | 249.133 | 315.000 | 277.680 | 287.625 | 295.486 | 127.925 | 253.240 | -56.7% | -14.3% |
| 265 | XL | 20 | Nội thành | 271.700 | 319.197 | 420.000 | 368.680 | 381.875 | 336.966 | 142.000 | 285.800 | -57.9% | -15.2% |
| 266 | XL | 20 | Ngoại thành | 271.700 | 319.197 | 420.000 | 368.680 | 381.875 | 336.966 | 142.000 | 285.800 | -57.9% | -15.2% |
| 267 | XL | 20 | Sân bay | 271.700 | 319.197 | 420.000 | 446.680 | 381.875 | 336.966 | 152.000 | 295.800 | -54.9% | -12.2% |
| 268 | XL | 20 | Giờ cao điểm | 271.700 | 319.197 | 420.000 | 405.548 | 381.875 | 387.510 | 156.000 | 314.080 | -59.7% | -18.9% |
| 269 | XL | 20 | Mưa | 271.700 | 319.197 | 420.000 | 368.680 | 381.875 | 387.510 | 163.000 | 328.220 | -57.9% | -15.3% |
| 270 | XL | 20 | Đêm | 271.700 | 319.197 | 420.000 | 368.680 | 381.875 | 387.510 | 170.000 | 342.360 | -56.1% | -11.7% |
| 271 | XL | 20 | Cuối tuần | 271.700 | 319.197 | 420.000 | 368.680 | 381.875 | 336.966 | 142.000 | 285.800 | -57.9% | -15.2% |
| 272 | XL | 20 | Lễ Tết | 271.700 | 319.197 | 420.000 | 368.680 | 381.875 | 387.510 | 163.000 | 328.220 | -57.9% | -15.3% |
| 273 | XL | 25 | Nội thành | 336.700 | 389.260 | 525.000 | 459.680 | 476.125 | 416.987 | 172.500 | 351.000 | -58.6% | -15.8% |
| 274 | XL | 25 | Ngoại thành | 336.700 | 389.260 | 525.000 | 459.680 | 476.125 | 416.987 | 172.500 | 351.000 | -58.6% | -15.8% |
| 275 | XL | 25 | Sân bay | 336.700 | 389.260 | 525.000 | 557.180 | 476.125 | 416.987 | 182.500 | 361.000 | -56.2% | -13.4% |
| 276 | XL | 25 | Giờ cao điểm | 336.700 | 389.260 | 525.000 | 505.648 | 476.125 | 479.535 | 189.550 | 385.800 | -60.5% | -19.5% |
| 277 | XL | 25 | Mưa | 336.700 | 389.260 | 525.000 | 459.680 | 476.125 | 479.535 | 198.075 | 403.200 | -58.7% | -15.9% |
| 278 | XL | 25 | Đêm | 336.700 | 389.260 | 525.000 | 459.680 | 476.125 | 479.535 | 206.600 | 420.600 | -56.9% | -12.3% |
| 279 | XL | 25 | Cuối tuần | 336.700 | 389.260 | 525.000 | 459.680 | 476.125 | 416.987 | 172.500 | 351.000 | -58.6% | -15.8% |
| 280 | XL | 25 | Lễ Tết | 336.700 | 389.260 | 525.000 | 459.680 | 476.125 | 479.535 | 198.075 | 403.200 | -58.7% | -15.9% |
| 281 | XL | 30 | Nội thành | 401.700 | 459.324 | 630.000 | 550.680 | 570.375 | 497.008 | 203.000 | 416.200 | -59.2% | -16.3% |
| 282 | XL | 30 | Ngoại thành | 401.700 | 459.324 | 630.000 | 550.680 | 570.375 | 497.008 | 203.000 | 416.200 | -59.2% | -16.3% |
| 283 | XL | 30 | Sân bay | 401.700 | 459.324 | 630.000 | 667.680 | 570.375 | 497.008 | 213.000 | 426.200 | -57.1% | -14.2% |
| 284 | XL | 30 | Giờ cao điểm | 401.700 | 459.324 | 630.000 | 605.748 | 570.375 | 571.559 | 223.100 | 457.520 | -61.0% | -20.0% |
| 285 | XL | 30 | Mưa | 401.700 | 459.324 | 630.000 | 550.680 | 570.375 | 571.559 | 233.150 | 478.180 | -59.2% | -16.3% |
| 286 | XL | 30 | Đêm | 401.700 | 459.324 | 630.000 | 550.680 | 570.375 | 571.559 | 243.200 | 498.840 | -57.4% | -12.7% |
| 287 | XL | 30 | Cuối tuần | 401.700 | 459.324 | 630.000 | 550.680 | 570.375 | 497.008 | 203.000 | 416.200 | -59.2% | -16.3% |
| 288 | XL | 30 | Lễ Tết | 401.700 | 459.324 | 630.000 | 550.680 | 570.375 | 571.559 | 233.150 | 478.180 | -59.2% | -16.3% |
| 289 | XL | 40 | Nội thành | 531.700 | 599.451 | 840.000 | 693.680 | 724.945 | 657.050 | 264.000 | **546.600** | -59.8% | -16.8% |
| 290 | XL | 40 | Ngoại thành | 531.700 | 599.451 | 840.000 | 693.680 | 724.945 | 657.050 | 264.000 | **546.600** | -59.8% | -16.8% |
| 291 | XL | 40 | Sân bay | 531.700 | 599.451 | 840.000 | 849.680 | 724.945 | 657.050 | 274.000 | **556.600** | -58.3% | -15.3% |
| 292 | XL | 40 | Giờ cao điểm | 531.700 | 599.451 | 840.000 | 763.048 | 724.945 | 755.608 | 290.200 | **600.960** | -61.6% | -20.5% |
| 293 | XL | 40 | Mưa | 531.700 | 599.451 | 840.000 | 693.680 | 724.945 | 755.608 | 303.300 | **628.140** | -59.9% | -16.9% |
| 294 | XL | 40 | Đêm | 531.700 | 599.451 | 840.000 | 693.680 | 724.945 | 755.608 | 316.400 | **655.320** | -58.1% | -13.3% |
| 295 | XL | 40 | Cuối tuần | 531.700 | 599.451 | 840.000 | 693.680 | 724.945 | 657.050 | 264.000 | **546.600** | -59.8% | -16.8% |
| 296 | XL | 40 | Lễ Tết | 531.700 | 599.451 | 840.000 | 693.680 | 724.945 | 755.608 | 303.300 | **628.140** | -59.9% | -16.9% |
| 297 | XL | 50 | Nội thành | 661.700 | 739.578 | 1.050.000 | 836.680 | 875.745 | 817.093 | 325.000 | **677.000** | -60.2% | -17.1% |
| 298 | XL | 50 | Ngoại thành | 661.700 | 739.578 | 1.050.000 | 836.680 | 875.745 | 817.093 | 325.000 | **677.000** | -60.2% | -17.1% |
| 299 | XL | 50 | Sân bay | 661.700 | 739.578 | 1.050.000 | 1.031.680 | 875.745 | 817.093 | 335.000 | **687.000** | -59.0% | -15.9% |
| 300 | XL | 50 | Giờ cao điểm | 661.700 | 739.578 | 1.050.000 | 920.348 | 875.745 | 939.656 | 357.300 | **744.400** | -62.0% | -20.8% |
| 301 | XL | 50 | Mưa | 661.700 | 739.578 | 1.050.000 | 836.680 | 875.745 | 939.656 | 373.450 | **778.100** | -60.3% | -17.2% |
| 302 | XL | 50 | Đêm | 661.700 | 739.578 | 1.050.000 | 836.680 | 875.745 | 939.656 | 389.600 | **811.800** | -58.5% | -13.6% |
| 303 | XL | 50 | Cuối tuần | 661.700 | 739.578 | 1.050.000 | 836.680 | 875.745 | 817.093 | 325.000 | **677.000** | -60.2% | -17.1% |
| 304 | XL | 50 | Lễ Tết | 661.700 | 739.578 | 1.050.000 | 836.680 | 875.745 | 939.656 | 373.450 | **778.100** | -60.3% | -17.2% |
| 305 | XL | 60 | Nội thành | 791.700 | 879.705 | 1.260.000 | 979.680 | 1.026.545 | 977.135 | 386.000 | **807.400** | -60.5% | -17.4% |
| 306 | XL | 60 | Ngoại thành | 791.700 | 879.705 | 1.260.000 | 979.680 | 1.026.545 | 977.135 | 386.000 | **807.400** | -60.5% | -17.4% |
| 307 | XL | 60 | Sân bay | 791.700 | 879.705 | 1.260.000 | 1.213.680 | 1.026.545 | 977.135 | 396.000 | **817.400** | -59.5% | -16.3% |
| 308 | XL | 60 | Giờ cao điểm | 791.700 | 879.705 | 1.260.000 | 1.077.648 | 1.026.545 | 1.123.705 | 424.400 | **887.840** | -62.2% | -21.0% |
| 309 | XL | 60 | Mưa | 791.700 | 879.705 | 1.260.000 | 979.680 | 1.026.545 | 1.123.705 | 443.600 | **928.060** | -60.5% | -17.4% |
| 310 | XL | 60 | Đêm | 791.700 | 879.705 | 1.260.000 | 979.680 | 1.026.545 | 1.123.705 | 462.800 | **968.280** | -58.8% | -13.8% |
| 311 | XL | 60 | Cuối tuần | 791.700 | 879.705 | 1.260.000 | 979.680 | 1.026.545 | 977.135 | 386.000 | **807.400** | -60.5% | -17.4% |
| 312 | XL | 60 | Lễ Tết | 791.700 | 879.705 | 1.260.000 | 979.680 | 1.026.545 | 1.123.705 | 443.600 | **928.060** | -60.5% | -17.4% |

**In đậm** = giá "Panda đề xuất" vượt trần giá tuyệt đối BRB §2.13.6 (500.000đ) — xem cảnh báo Phần 8.4, cần xử lý trước khi duyệt chính thức.

---

## PHẦN 5 — THU NHẬP TÀI XẾ (waterfall)

**Giả định chi phí (ASSUMPTION, không có trong bất kỳ tài liệu nào đã đọc — cần CFO xác nhận):** xăng 25.000đ/lít; tiêu hao bike 2.0L/100km, car 8.0L/100km, XL 9.5L/100km; khấu hao + bảo trì: bike 500đ/km, car 1.800đ/km, XL 2.200đ/km (ước lượng dựa trên vòng đời phương tiện điển hình tại Việt Nam, **không** phải số liệu kế toán thật của Panda).

Waterfall: **Khách trả → (Voucher, nếu có — xem Phần 6) → Commission (Bronze 20% / Diamond 12%, BRB §7.1) → Net Driver (có sàn Minimum Driver Earning 20.000đ, BRB §2.14) → trừ Xăng → trừ Khấu hao → Lợi nhuận ròng tài xế.**

### 5.1 Giá hiện tại (Bronze 20%, "Nội thành")

| Xe | Km | Khách trả | Commission | Net Driver | Xăng | Khấu hao | **Lợi nhuận ròng tài xế** |
|---|---|---|---|---|---|---|---|
| Bike | 5 | 22.640 | 4.128 | 20.000 (sàn) | 2.500 | 2.500 | **15.000** |
| Bike | 20 | 66.560 | 12.912 | 51.648 | 10.000 | 10.000 | **31.648** |
| Bike | 60 | 183.680 | 36.336 | 145.344 | 30.000 | 30.000 | **85.344** |
| Car | 5 | 36.400 | 6.880 | 27.520 | 10.000 | 9.000 | **8.520** |
| Car | 20 | 109.600 | 21.520 | 86.080 | 40.000 | 36.000 | **10.080** |
| Car | 40 | 207.200 | 41.040 | 164.160 | 80.000 | 72.000 | **12.160** |
| Car | 60 | 304.800 | 60.560 | 242.240 | 120.000 | 108.000 | **14.240** |
| XL | 5 | 50.500 | 9.700 | 38.800 | 11.875 | 11.000 | **15.925** |
| XL | 20 | 142.000 | 28.000 | 112.000 | 47.500 | 44.000 | **20.500** |
| XL | 60 | 386.000 | 76.800 | 307.200 | 142.500 | 132.000 | **32.700** |

**Phát hiện quan trọng nhất của toàn tài liệu:** với hạng Car, lợi nhuận ròng tài xế **gần như không tăng theo quãng đường** (8.520đ ở 5km → chỉ 14.240đ ở 60km, tăng 67% trong khi quãng đường tăng 1.100%) — vì /km hiện tại (4.000đ) gần bằng đúng chi phí xăng+khấu hao/km (3.800đ: 2.000đ xăng + 1.800đ khấu hao). Gần như **toàn bộ phần thu theo km chỉ đủ trả chi phí vận hành xe**, tài xế Car sống chủ yếu nhờ Base Fare + Booking Fee cố định, không nhờ quãng đường — một cấu trúc không khuyến khích tài xế nhận cuốc dài.

### 5.2 Giá đề xuất (Bronze 20%, "Nội thành")

| Xe | Km | Khách trả | Commission | Net Driver | Xăng | Khấu hao | **Lợi nhuận ròng tài xế** |
|---|---|---|---|---|---|---|---|
| Bike | 5 | 23.920 | 4.584 | 20.000 (sàn) | 2.500 | 2.500 | **15.000** |
| Bike | 20 | 85.180 | 16.836 | 67.344 | 10.000 | 10.000 | **47.344** |
| Bike | 60 | 248.540 | 49.508 | 198.032 | 30.000 | 30.000 | **138.032** |
| Car | 5 | 64.940 | 12.388 | 49.552 | 10.000 | 9.000 | **30.552** |
| Car | 20 | 211.760 | 41.752 | 167.008 | 40.000 | 36.000 | **91.008** |
| Car | 40 | 407.520 | 80.904 | 323.616 | 80.000 | 72.000 | **171.616** |
| Car | 60 | 603.280 | 120.056 | 480.224 | 120.000 | 108.000 | **252.224** |
| XL | 5 | 90.200 | 17.440 | 69.760 | 11.875 | 11.000 | **46.885** |
| XL | 20 | 285.800 | 56.560 | 226.240 | 47.500 | 44.000 | **134.740** |
| XL | 60 | 807.400 | 160.880 | 643.520 | 142.500 | 132.000 | **369.020** |

→ Với giá đề xuất, lợi nhuận ròng tài xế Car **tăng gần đúng tuyến tính theo km** (30.552đ ở 5km → 252.224đ ở 60km, tăng ~726% khi quãng đường tăng 1.100% — vẫn không hoàn toàn tuyến tính vì Base Fare/Booking Fee cố định, nhưng đúng hướng khỏe mạnh hơn nhiều so với 5.1).

### 5.3 Ở tier Diamond (12% hoa hồng) so với Bronze (20%) — cùng giá đề xuất

| Xe | Km | Net Driver (Bronze) | Net Driver (Diamond) | Chênh lệch |
|---|---|---|---|---|
| Car | 20 | 167.008 | 183.709 | +10.0% |
| Car | 60 | 480.224 | 528.246 | +10.0% |

→ Đúng thiết kế BRB §7.1: khoảng cách Bronze→Diamond luôn là **+10% thu nhập tuyệt đối trên cùng ride fare** bất kể giá gốc — xác nhận Commission Engine hoạt động độc lập với việc hiệu chỉnh giá (hai lớp không xung đột, đúng ECONOMY_ENGINE §5).

---

## PHẦN 6 — DOANH THU / LỢI NHUẬN PANDA MỖI CUỐC

**Giả định chi phí vận hành công nghệ/cuốc (ASSUMPTION — không có trong BRB/ECONOMY_ENGINE, cần CFO xác nhận):** phí cổng thanh toán ~1.8% giá trị giao dịch (MDR/QR phổ biến thị trường VN), SMS/notification ~80đ, map/routing API ~250đ, cloud/server phân bổ ~300đ, support/fraud provisioning ~250đ — tổng chi phí cố định/cuốc **~880đ + 1.8%×Khách trả**.

### 6.1 Bảng lợi nhuận nền tảng/cuốc (Bronze 20%, "Nội thành")

| Xe | Km | Rates | Khách trả | Commission+Booking | VAT (10%) | Phí gateway | Opex cố định | **Lợi nhuận Panda** | **Biên lợi nhuận** |
|---|---|---|---|---|---|---|---|---|---|
| Car | 5 | Hiện tại | 36.400 | 8.880 | 888 | 655 | 880 | **6.457** | 17.7% |
| Car | 5 | Đề xuất | 64.940 | 15.388 | 1.539 | 1.169 | 880 | **11.800** | 18.2% |
| Car | 20 | Hiện tại | 109.600 | 23.520 | 2.352 | 1.973 | 880 | **18.315** | 16.7% |
| Car | 20 | Đề xuất | 211.760 | 44.752 | 4.475 | 3.812 | 880 | **35.585** | 16.8% |
| Car | 60 | Hiện tại | 304.800 | 62.560 | 6.256 | 5.486 | 880 | **49.938** | 16.4% |
| Car | 60 | Đề xuất | 603.280 | 123.056 | 12.306 | 10.859 | 880 | **99.011** | 16.4% |
| Car | 60 | Đề xuất, tier Diamond | 603.280 | 75.034 | 7.503 | 10.859 | 880 | **55.791** | 9.2% |

**Nhận xét:** biên lợi nhuận Panda/cuốc ở Bronze **gần như không đổi (~16-18%)** dù giá hiện tại hay giá đề xuất — vì cấu trúc phí (VAT 10% trên doanh thu nền tảng, gateway 1.8% trên khách trả, opex cố định nhỏ so với giá trị giao dịch) đều tỷ lệ thuận với giá. Điều **thay đổi thật sự** không phải biên % mà là **số tuyệt đối**: cùng biên 16-18%, giá đề xuất mang lại lợi nhuận tuyệt đối gấp ~2 lần giá hiện tại cho cùng quãng đường — đây chính là khoản "dư địa" bị bỏ lỡ khi định giá dưới thị trường 42.5%.

Ở tier Diamond (nhiều tài xế lâu năm, hoa hồng chỉ 12%), biên lợi nhuận nền tảng giảm xuống **~9-11%** — vẫn dương, nhưng mỏng hơn đáng kể. Đây là lý do PRICING_SIMULATION_REPORT (sprint trước) khuyến nghị cân nhắc hạ Bronze xuống 16-18% để cân bằng tốt hơn giữa "tài xế thắng đối thủ" và "biên lợi nhuận nền tảng" — khuyến nghị đó **vẫn đúng và độc lập** với việc hiệu chỉnh giá ở tài liệu này (hai đòn bẩy khác nhau: mức giá gốc vs. tỷ lệ chia commission).

---

## PHẦN 7 — ĐIỂM HOÀ VỐN

### 7.1 Hoà vốn mỗi cuốc (per-trip contribution margin) — đã đạt, ở cả hai bảng giá

Theo Phần 6, biên lợi nhuận mỗi cuốc dương (~16-18% ở Bronze) ở **cả giá hiện tại lẫn giá đề xuất** — ngoại trừ các trường hợp loss-leader **có chủ đích** đã biết từ PRICING_SIMULATION_REPORT (voucher sâu, chuyến bike ≤2km do Minimum Driver Earning Guarantee, Long Pickup Compensation) — những trường hợp này **không phải** dấu hiệu mô hình giá sai, mà là chi phí tăng trưởng đã được BRB/PRICING_STRATEGY chấp nhận có kiểm soát (ngân sách riêng, BRB §3.3).

→ **Kết luận:** vấn đề "không bền vững" mà nhiệm vụ này xuất phát từ đó **không nằm ở việc mỗi cuốc đang lỗ** (thực tế không lỗ) — nó nằm ở việc **giá quá thấp so với thị trường khiến biên an toàn tuyệt đối (VND/cuốc) quá mỏng để chống chịu cú sốc chi phí** (xem Phần 10 — cú sốc xăng +20% đẩy lợi nhuận tài xế Car về âm ở giá hiện tại nhưng không ở giá đề xuất).

### 7.2 Hoà vốn toàn nền tảng (company-wide, có chi phí cố định) — ước lượng minh bạch, có giả định

**Không có dữ liệu thật** về chi phí cố định hàng tháng (nhân sự, văn phòng, marketing, hạ tầng nền) trong bất kỳ tài liệu nào đã đọc — PRICING_STRATEGY chỉ đặt mục tiêu định tính ("hoà vốn ở Giai đoạn 4, 10.000 tài xế, 12 tháng", §4.5). Để minh hoạ, tài liệu này dùng **một giả định duy nhất, công bố rõ, không phải số thật**: chi phí cố định ~1 tỷ VND/tháng (đội ngũ giai đoạn seed ~15-20 người + văn phòng + hạ tầng nền + marketing tối thiểu).

- Lợi nhuận trung bình/cuốc (biên ~17%, quy mô trung bình ~57.000đ ở giá đề xuất theo Phần 4) ≈ **~9.700đ/cuốc** đóng góp cho chi phí cố định.
- Số cuốc/tháng cần để hoà vốn công ty ≈ 1.000.000.000 / 9.700 ≈ **~103.000 cuốc/tháng** ≈ **~3.400 cuốc/ngày**.
- Đối chiếu PRICING_STRATEGY §4.5 (Giai đoạn 4, 10.000 tài xế): nếu mỗi tài xế chạy trung bình ~10-15 cuốc/ngày (giả định phổ biến ngành), 10.000 tài xế tạo ra ~100.000-150.000 cuốc/ngày — **vượt xa** ngưỡng hoà vốn ước lượng ở trên. Điều này gợi ý mục tiêu định tính "hoà vốn ở Giai đoạn 4" trong PRICING_STRATEGY là **khả thi về mặt số học**, miễn là giá thực sự được hiệu chỉnh lên gần bảng đề xuất (ở giá hiện tại -42.5%, lợi nhuận/cuốc thấp hơn đáng kể, cần nhiều cuốc hơn để đạt cùng ngưỡng).

**Cảnh báo minh bạch:** đây là một phép tính bậc-độ-lớn (order-of-magnitude) dựa trên MỘT giả định chi phí cố định chưa được CFO xác nhận — không dùng làm cam kết ngân sách hay dự báo tài chính chính thức.

---

## PHẦN 8 — ĐỀ XUẤT BẢNG GIÁ MỚI (đồng bộ)

### 8.1 Mục tiêu định vị (không phải "rẻ nhất")

Đúng định vị "Fairest & Most Transparent, không phải Cheapest" đã chọn tại PRICING_STRATEGY §1: mục tiêu là rẻ hơn thị trường **một khoảng vừa đủ để là lựa chọn hợp lý khi so giá** (PS §1 ưu tiên #4), không phải đáy giá.

| Hạng xe | Mục tiêu | Kết quả mô phỏng (Phần 4) | Đạt? |
|---|---|---|---|
| Bike | 8-12% rẻ hơn thị trường | -9.8% | ✓ |
| Car | 10-15% rẻ hơn thị trường | -13.5% | ✓ |
| XL | 8-12% rẻ hơn thị trường | -11.0% | ✓ |
| Với Promotion áp dụng | Tối đa ~20% rẻ hơn thị trường (chưa tính campaign sâu như First Ride) | Không đổi — Promotion là lớp riêng, không đổi giá niêm yết (Phần 1) | — |

### 8.2 Bảng giá đề xuất (VND, thiết kế đồng bộ — không chỉnh từng số riêng lẻ)

| Thành phần | Car (Standard) | XL | Bike |
|---|---|---|---|
| Base Fare | 13.000 | 22.000 | 2.500 |
| /km | 8.600 | 11.500 | 3.600 |
| /phút | 540 | 700 | 230 |
| Minimum Fare | 30.000 | 48.000 | 9.000 |
| Booking Fee | 3.000 | 3.000 | 1.000 |

Giữ nguyên toàn bộ cơ chế phụ phí đã có (Airport Fee, Night/Holiday/Rain/Peak multiplier, trần cộng dồn ×1.60, Waiting Fee, Long Pickup Compensation, trần surge ×2.0) — bảng trên **chỉ thay 5 tham số gốc**, không đổi công thức/luật đã có trong BRB §2.2, đúng nguyên tắc "thiết kế đồng bộ, không vá từng số" của nhiệm vụ.

### 8.3 Vì sao Booking Fee cũng tăng (3.000đ thay vì 2.000đ)

Khi Base/km/phút được nâng lên gần thị trường, Booking Fee giữ nguyên 2.000đ sẽ khiến vai trò "phí dịch vụ nền tảng" của nó bị pha loãng tương đối (từ ~7% giá chuyến ngắn hiện tại xuống dưới 3% giá chuyến ngắn nếu các thành phần khác tăng mà Booking Fee đứng yên) — nâng nhẹ lên 3.000đ giữ đúng tỷ trọng ban đầu, đây là một thay đổi **[MỚI — cần tu chính BRB §2.2.5]**, không phải điều chỉnh tuỳ tiện.

### 8.4 CẢNH BÁO — cần xử lý trước khi trình duyệt chính thức

Bảng đề xuất ở dạng **phẳng** (một mức /km duy nhất, không giảm dần theo quãng đường) trong khi **cả 5 đối thủ nghiên cứu đều có bậc giảm giá quãng đường dài** (Grab/Be/GreenSM/Mai Linh/Vinasun, xem Phần 2). Hệ quả: ở quãng đường ≥ 50km, giá Car/XL đề xuất **vượt trần giá tuyệt đối BRB §2.13.6 (500.000 VND)** — xem các dòng in đậm trong bảng Phần 4.3 (ví dụ Car 50km/nội thành = 505.400đ, Car 60km = 603.280đ).

**Khuyến nghị cụ thể (chưa áp dụng, cần quyết định CPO/CFO):**
1. Thêm bậc giảm /km sau 30km (ví dụ: -15% so với /km gốc), mô phỏng lại toàn bộ 312 kịch bản trước khi trình duyệt — mirror đúng pattern cả 5 đối thủ đã có.
2. Hoặc áp dụng đúng ghi chú đã có sẵn trong BRB §2.13.6 ("Airport long-distance trips have a higher cap negotiated separately") — mở rộng ngoại lệ này cho **mọi** chuyến dài ≥ 40km, không chỉsân bay, vì bản thân thị trường (screenshots Be/GreenSM Phần 2) cho thấy chuyến 50-60km của mọi nền tảng đều tự nhiên vượt 500.000đ.
3. Việc này **không** làm vô hiệu kết quả mô phỏng Phần 4/5/6/7 cho quãng đường < 40km (chiếm phần lớn volume chuyến thực tế theo trực giác vận hành đô thị) — chỉ ảnh hưởng phần đuôi phân phối (chuyến rất dài/liên tỉnh).

---

## PHẦN 9 — MARKET INDEX

### 9.1 Thiết kế

Thay vì Pricing Service tự tính lại Base/km/phút mỗi khi cần điều chỉnh vị thế cạnh tranh, đề xuất một **hằng số cấu hình duy nhất mỗi hạng xe — TargetIndex** — biểu diễn vị thế giá của Panda so với **Market Reference** (trung bình 3 nền tảng công nghệ nghiên cứu ở Phần 2, cùng cách tính đã dùng xuyên suốt tài liệu này).

```
Panda_Fare(scenario) = ReferenceMarketFare(scenario) × TargetIndex[hạng xe]
```

**Quan trọng — đây KHÔNG phải một lời gọi API sống tới Grab/Be/GreenSM lúc runtime** (vi phạm nguyên tắc Route Engine độc lập của Panda, BRB §2.2.2, và tạo rủi ro phụ thuộc bên thứ ba y hệt bài học Yandex Go đã ghi trong PRICING_STRATEGY §0). `ReferenceMarketFare` là một **đường cong tham chiếu tĩnh**, hiệu chỉnh **định kỳ** (ví dụ hàng quý, do đội Pricing/Finance khảo sát thị trường thủ công — đúng cách tài liệu này đã làm) và lưu thành cấu hình (Base/km/phút riêng của "Market Reference"), không phải một dependency thời gian thực. `TargetIndex` áp lên đường cong tham chiếu này để suy ra Base/km/phút/Minimum Fare thật của Panda (đúng phương pháp Phần 8.2 đã dùng để dò ra 5 tham số).

### 9.2 Chỉ số thật đo được (không phải ví dụ minh hoạ)

Khác với ví dụ minh hoạ trong đề bài nhiệm vụ (Grab=100%, Be=98%, GreenSM=95% — giả định Be/GreenSM rẻ hơn Grab), **số liệu nghiên cứu thật (Phần 2) cho kết quả ngược lại đối với hạng Car**:

| Nền tảng | Index so với Grab (=100%, hạng Car, TB 13 khoảng cách) |
|---|---|
| Grab | 100% (mốc neo) |
| Be | ~123% |
| GreenSM | ~138% |
| **Panda hiện tại** | ~47% |
| **Panda đề xuất** | ~87% |

(Với Bike: Grab 100%, Be ~105%, GreenSM ~108% — ba nền tảng gần nhau hơn nhiều so với Car.)

**Tài liệu này giữ số đo thật thay vì ép theo ví dụ minh hoạ của đề bài** — đúng nguyên tắc không bịa số của toàn bộ hệ thống tài liệu Panda (never fabricate).

### 9.3 TargetIndex đề xuất

| Hạng xe | TargetIndex (so với Market Reference = TB Grab+Be+GreenSM) |
|---|---|
| Car | 0.855-0.90 (tương ứng -10% đến -15%) |
| XL | 0.88-0.92 (tương ứng -8% đến -12%) |
| Bike | 0.88-0.92 (tương ứng -8% đến -12%) |

Vận hành: Pricing Service tương lai chỉ cần đọc `TargetIndex[vehicle]` + `ReferenceMarketFare[vehicle]` (bảng cấu hình, cập nhật định kỳ) để tính ra `VehicleRates` — không đổi kiến trúc `FareConfig` hiện có, chỉ thêm một lớp cấu hình phía trên nó (đúng nguyên tắc Rule Engine "không hardcode" của ECONOMY_ENGINE Phần 11).

---

## PHẦN 10 — SENSITIVITY (bảng giá đề xuất có còn sống không?)

| Cú sốc | Giá hiện tại | Giá đề xuất | Kết luận |
|---|---|---|---|
| **Xăng +20%** (car 40km) | Lợi nhuận ròng tài xế: 12.160đ → **-3.840đ (ÂM)** | 171.616đ → 155.616đ (-9.3%, vẫn dương mạnh) | ⚠️ Giá hiện tại **không sống nổi** một cú sốc xăng vừa phải cho tài xế Car ở quãng trung-dài. Giá đề xuất có đủ đệm |
| **Xăng +20%** (car 60km) | 14.240đ → **-9.760đ (ÂM)** | 252.224đ → 228.224đ (-9.5%, vẫn dương mạnh) | Cùng kết luận — càng chuyến dài, giá hiện tại càng dễ âm |
| **Driver tăng gấp đôi (DSR giảm 1/2)** | Kế thừa nguyên vẹn kết quả đã có tại `PRICING_SIMULATION_REPORT.md` Phần 6: surge trung bình 1.58x→1.14x, thu nhập tài xế từ surge -28.4% | Không đổi so với giá hiện tại — cơ chế Surge (hệ số nhân) độc lập với mức giá gốc, chỉ ảnh hưởng bởi DSR | ✅ Không có rủi ro mới do hiệu chỉnh giá — đây là cơ chế thị trường đúng thiết kế |
| **Khách tăng gấp 5 (DSR ×5)** | Kế thừa `PRICING_SIMULATION_REPORT.md`: surge chạm trần ×2.0 ở toàn bộ scenario có surge; trần giá tuyệt đối chưa bị chạm ở giá cũ | **Trần giá tuyệt đối 500.000đ đã bị chạm ngay cả KHÔNG có surge** ở quãng ≥50km (xem Phần 8.4) — cộng thêm surge ×2.0 sẽ vượt trần rất xa | ⚠️ Xác nhận lại phát hiện Phần 8.4 — cần bậc giảm giá quãng đường dài hoặc nới trần trước khi surge được bật lên giá đề xuất |
| **Voucher/Promotion tăng (chiết khấu sâu 20%, theo Phần 7 mục tiêu tối đa)** | Kế thừa `PRICING_SIMULATION_REPORT.md`: đây là loss-leader có chủ đích, lỗ tăng tuyến tính theo ngân sách, không phi mã | Cùng cơ chế — một cuốc Car 20km giảm 20% (ví dụ minh hoạ): lợi nhuận nền tảng từ +35.585đ chuyển thành **âm ~-6.200đ cho riêng cuốc đó**, được bù bằng Promotion Fund (ngân sách riêng, BRB §3.3), không phải từ doanh thu thường xuyên | ✅ Đúng thiết kế đã có — không phải rủi ro mới, miễn ngân sách Promotion Fund được kiểm soát đúng BRB §3.3 |
| **Commission giảm 20%→16% (khuyến nghị PRICING_SIMULATION_REPORT chưa áp dụng)** | Không đổi kết luận sprint trước | Cùng hướng: biên lợi nhuận nền tảng giảm nhẹ, thu nhập tài xế tăng — vẫn dương ở cả hai bảng giá (Phần 6 cho thấy biên Diamond 12% vẫn +9.2% ở giá đề xuất) | ✅ An toàn ở cả hai bảng giá |

**Kết luận Phần 10:** bảng giá đề xuất **giải quyết đúng lỗ hổng nghiêm trọng nhất** (chi phí xăng/khấu hao ăn hết thu nhập theo km ở giá hiện tại) nhưng **tạo ra một lỗ hổng mới cần xử lý trước khi duyệt** (vượt trần giá tuyệt đối ở quãng dài — Phần 8.4). Không có sốc nào trong 5 sốc trên phá vỡ tính khả thi của giá đề xuất **sau khi** Phần 8.4 được xử lý.

---

## PHẦN 11 — ROADMAP CHUYỂN ĐỔI & FILES CẦN SỬA SAU KHI DUYỆT

### 11.1 Roadmap chuyển đổi (không tự áp dụng — chờ CPO/CFO/CTO phê duyệt)

| Bước | Nội dung | Điều kiện tiên quyết |
|---|---|---|
| 1 | CPO/CFO/CTO đọc và phê duyệt tài liệu này | — |
| 2 | Xử lý cảnh báo Phần 8.4 (bậc giảm giá quãng đường dài hoặc nới trần) — mô phỏng lại 312 kịch bản với bậc mới | Bước 1 |
| 3 | Tu chính BRB v1.1: đổi 5 tham số Base/km/phút/Minimum/Booking (Phần 8.2, sau khi sửa Phần 8.4), thêm bậc giảm giá quãng dài, thêm Market Index framework (Phần 9) như một Part mới | Bước 2 + quy trình tu chính chính thức Constitution Article XI |
| 4 | Đồng bộ `fare.go` (production) và `simulation/pricing_constants.go` (Phần 1.4 — tỷ lệ xe máy 0.60 vs ~0.5) | Bước 3 |
| 5 | Quyết định CPO về Airport Fee theo hạng xe (Phần 1.3 — có nên miễn Airport Fee cho Bike hay không) | Bước 3 |
| 6 | Cân nhắc khuyến nghị hạ Bronze commission 20%→16-18% từ `PRICING_SIMULATION_REPORT.md` (độc lập với bảng giá mới, vẫn chưa áp dụng) | Bước 3, quyết định riêng |
| 7 | Chạy lại toàn bộ test Go hiện có (`fare_calculator_test.go`, `fare_calculator_dynamic_test.go`, `handler_test.go` — đã cập nhật theo BRB VND ở phiên trước) với tham số mới | Bước 4 |
| 8 | Cập nhật `MockBookingCatalog`/`MockFareBreakdown` (rider app) theo tham số mới để giá ước tính trước khi đặt xe khớp giá thật | Bước 4 |
| 9 | CFO xác nhận chính thức các ASSUMPTION đã dùng xuyên suốt tài liệu (VAT 10%, chi phí xăng/khấu hao Phần 5, chi phí opex/cuốc Phần 6, chi phí cố định Phần 7) — mọi con số chưa xác nhận **không được** dùng làm cơ sở quyết định cuối cùng | Song song bước 3 |
| 10 | Rollout theo giai đoạn tăng trưởng đã có tại PRICING_STRATEGY §4 (không đổi khung giai đoạn, chỉ thay số tuyệt đối bên trong mỗi giai đoạn) | Bước 3 |

### 11.2 Files cần sửa (chỉ sau khi được duyệt chính thức — KHÔNG sửa ngay)

| File | Thay đổi cần thiết |
|---|---|
| `docs/business/business-rule-bible-v1.0.md` | Tu chính §2.2.1-§2.2.5 theo bảng giá mới (sau khi xử lý Phần 8.4); thêm mục Market Index (Phần 9) như Part mới; thêm mục bậc giảm giá quãng đường dài |
| `backend/services/pricing/domain/entity/fare.go` | `DefaultFareConfig()` — cập nhật 5 tham số mỗi hạng xe theo bảng mới; cân nhắc thêm cấu trúc bậc quãng đường dài (thay đổi shape, không chỉ giá trị) |
| `backend/services/pricing/simulation/pricing_constants.go` | Đồng bộ `MotorcycleFareRatio` với quyết định cuối cùng (Phần 1.4); cập nhật rates khớp BRB mới |
| `backend/services/pricing/app/fare_calculator_test.go`, `fare_calculator_dynamic_test.go` | Cập nhật số kỳ vọng theo rates mới (đã có tiền lệ từ phiên trước — cùng phương pháp) |
| `backend/services/pricing/grpc/handler_test.go` | Cùng như trên |
| `apps/rider/lib/features/booking/domain/models/mock_booking_catalog.dart` | Cập nhật rates ước tính trước khi đặt xe khớp bảng mới |
| `apps/rider/lib/features/booking/domain/models/mock_fare_calculator.dart` | Không đổi công thức, chỉ đổi số nguồn từ `mock_booking_catalog.dart` |
| (Mới) `backend/services/pricing/domain/entity/fare_market_index.go` hoặc tương đương | Nếu Market Index (Phần 9) được duyệt triển khai — bảng cấu hình `ReferenceMarketFare` + `TargetIndex`, tách biệt khỏi `DefaultFareConfig` hiện có |
| CHANGELOG.md | Ghi nhận thay đổi bảng giá sau khi triển khai thật (không phải tài liệu nghiên cứu này) |

---

## GHI CHÚ PHƯƠNG PHÁP (minh bạch, không giấu giếm)

- Toàn bộ 312 kịch bản được tính bằng script Python viết riêng cho tài liệu này (`pricing_research.py`, `driver_platform_economics.py`) — không phải một phần bàn giao production, không nằm trong `backend/services/pricing`, chỉ dùng để sinh số liệu cho báo cáo này, tương tự cách sprint trước dùng bản port Node.js cho `PRICING_SIMULATION_REPORT.md`.
- Rate card đối thủ: nghiên cứu công khai qua WebSearch ngày 2026-07-11, đã trích dẫn nguồn ở Phần 2 — **không phải** dữ liệu nội bộ của Grab/Be/Xanh SM/Mai Linh/Vinasun, có thể lệch so với giá thật tại một thời điểm/khu vực cụ thể.
- Mọi con số không có nguồn công khai đều đánh dấu **ASSUMPTION** rõ ràng tại chỗ xuất hiện (chi phí xăng/khấu hao Phần 5, chi phí vận hành công nghệ Phần 6, chi phí cố định Phần 7, hệ số XL +30%, phụ phí sân bay/surge giả định +15% cho đối thủ công nghệ ở điều kiện đặc biệt) — **không** được diễn giải thành số liệu đã xác nhận.
- Bảng giá đề xuất (Phần 8.2) đạt đúng mục tiêu trung bình theo hạng xe (Phần 4.1), nhưng **có phương sai theo khoảng cách/điều kiện** (xem cột Δ đề xuất trong bảng 312 dòng — dao động rộng hơn ở Sân bay/Giờ cao điểm do cách cộng phụ phí khác nhau giữa Panda và giả định đối thủ) — đây là hạn chế cố hữu của việc dùng MỘT bộ tham số cố định để khớp NHIỀU đường cong đối thủ có hình dạng khác nhau, không phải lỗi tính toán.
- Tài liệu này **kế thừa, không thay thế** `PRICING_SIMULATION_REPORT.md` (111 scenario, sprint trước) — mọi phát hiện của tài liệu đó (chuyến bike ngắn lỗ có cấu trúc, Long Pickup Compensation tốn kém, khuyến nghị hạ Bronze commission) vẫn còn nguyên giá trị và **không** bị tài liệu này thay đổi hay phủ định.

---

*Kết thúc tài liệu — Panda Market Pricing Research — v0.1 (Nghiên cứu).*
*Không sửa Business Rule Bible. Không sửa Pricing Service. Không sửa UI. Không build. Không commit.*
*Tài liệu này phải được CPO, CFO, CTO phê duyệt trước khi bất kỳ con số nào ở Phần 8/9 được đưa vào quy trình tu chính BRB chính thức.*
