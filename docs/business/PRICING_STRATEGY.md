# Panda — Chiến lược Định giá Tổng thể (Master Pricing Strategy)

**Document Classification:** Internal — Confidential
**Authority:** CEO · CPO · CFO · COO
**Effective Date:** 2026-07-10
**Status:** Draft v0.1 — chờ phê duyệt
**Review Cycle:** Hàng quý trong 3 năm đầu; hàng năm sau đó
**Supersedes:** Không có (tài liệu chiến lược giá đầu tiên)
**Tài liệu liên quan:**
- `docs/business/mission/project-constitution-v0.1.md` — Hiến pháp kỹ thuật (Article I §1.2 "Fairness First" là nền tảng triết lý của tài liệu này)
- `docs/business/business-rule-bible-v1.0.md` (viết tắt **BRB** trong toàn bộ tài liệu này) — cuốn "kinh thánh" quy tắc vận hành: công thức cước phí chính xác, bảng hoa hồng theo tier, toàn bộ Incentive Engine, Promotion Engine, Wallet, Settlement

> **Ghi chú về thương hiệu:** các tài liệu EOS gốc (Constitution, BRB) được soạn dưới tên dự án kỹ thuật "FAIRRIDE". Sản phẩm thực tế đã ra mắt mang thương hiệu **Panda** (rider) và **PandaDriver** (driver). Tài liệu này dùng "Panda" xuyên suốt vì đó là tên thương hiệu người dùng thực sự nhìn thấy — mọi nguyên tắc của FAIRRIDE Constitution và BRB áp dụng nguyên vẹn cho Panda.

> **Quan hệ với Business Rule Bible:** tài liệu này là tầng **CHIẾN LƯỢC** ("tại sao" và "theo trình tự nào") — nó không lặp lại các công thức cơ học đã có trong BRB. Mọi con số cước phí, hoa hồng, khuyến mãi chi tiết trong BRB vẫn là **Single Source of Truth (SSOT)** cho tầng vận hành ("chính xác bao nhiêu"), đúng nguyên tắc SSOT tại Constitution Article III §3.2. Khi tài liệu này đề xuất một quy tắc mới chưa có trong BRB (đánh dấu **[MỚI — cần tu chính BRB]**), quy tắc đó phải đi qua quy trình tu chính chính thức trước khi triển khai.

---

## TÓM TẮT ĐIỀU HÀNH

Panda **không** định giá để tối đa hoá lợi nhuận trong 3 năm đầu. Panda định giá theo một **thang ưu tiên 5 bậc** cố định:

1. **Tăng số lượng Rider** — không có người đặt xe, không có gì để bàn.
2. **Tăng số lượng Driver** — không có tài xế, rider đặt xe không ai nhận.
3. **Tăng tần suất sử dụng** — một rider/driver dùng 1 lần/tháng không tạo ra một nền tảng; dùng 10 lần/tháng thì có.
4. **Tăng tính cạnh tranh** — Panda phải là lựa chọn hợp lý khi người dùng mở 3 app cùng lúc để so giá.
5. **Hoà vốn** — chi phí vận hành phải được trang trải trước khi nghĩ đến lợi nhuận.

**Lợi nhuận đứng thứ 6, không có trong danh sách 5 bậc trên, và chỉ được cân nhắc chính thức từ Giai đoạn 5 (100,000 người dùng)** — xem PHẦN 4 và PHẦN 9.

Triết lý được chọn: **"Công bằng & Minh bạch, thực thi qua trình tự ưu tiên Driver-Enabled Growth"** (chi tiết PHẦN 1). Đây không phải khẩu hiệu — nó khớp trực tiếp với nguyên tắc "Fairness First" bất di bất dịch tại Constitution Article XI §11.3 và triết lý "Driver-First" tại BRB §1.4, nghĩa là tài liệu này **không tạo ra một triết lý mới**, nó **vận hành hoá** triết lý đã tồn tại thành một chiến lược giá theo giai đoạn cụ thể.

Panda **không copy** mô hình của Grab, Be, Xanh SM, Uber, Lyft, DiDi, Bolt, inDrive, Maxim hay Yandex Go. Panda học có chọn lọc từ cả 10 (PHẦN 0) và từ chối rõ ràng các yếu tố không phù hợp: hoa hồng cao và thuật toán mờ (Grab, Uber), đốt tiền không đo ROI (DiDi), mô hình đấu giá gây khó chịu (inDrive), hy sinh an toàn để giữ giá rẻ (Maxim), mô hình sở hữu tài sản nặng vốn (Xanh SM).

---

# PHẦN 0 — NGHIÊN CỨU THỊ TRƯỜNG: MÔ HÌNH GIÁ CỦA 10 ĐỐI THỦ

## 0.1 Phương pháp

Phân tích dưới đây dựa trên đặc điểm mô hình kinh doanh **công khai và đã được biết đến rộng rãi** của từng nền tảng — định vị chiến lược, cấu trúc hoa hồng điển hình, và cách vận hành khuyến mãi/surge. Đây **không phải** là số liệu tài chính chính xác tại một thời điểm (các con số thay đổi theo thị trường và thời gian, và không nền tảng nào công bố đầy đủ) — mục đích là rút ra **bài học chiến lược**, không phải sao chép số liệu.

## 0.2 Bảng so sánh

| Nền tảng | Đặc trưng mô hình giá | Điểm mạnh | Điểm yếu |
|---|---|---|---|
| **Grab** | Siêu ứng dụng đa dịch vụ (ride + food + pay + ads); surge thuật toán phức tạp; hoa hồng thuộc nhóm cao nhất khu vực | Mạng lưới tài xế khổng lồ, dữ liệu định giá tinh vi, thương hiệu số 1 Đông Nam Á, hệ sinh thái chéo dịch vụ (subsidize ride bằng lợi nhuận ads/fintech) | Hoa hồng cao gây bất mãn tài xế kéo dài; thuật toán surge/matching không minh bạch với tài xế; rider cảm nhận giá đắt hơn taxi truyền thống vào giờ cao điểm |
| **Be** | Định vị "thương hiệu Việt", tập trung thị trường nội địa; khuyến mãi tài xế mạnh để giành thị phần; tích hợp beFinancial | Bản sắc địa phương rõ ràng là lợi thế cạnh tranh thật; linh hoạt phản ứng thị trường trong nước nhanh hơn đối thủ ngoại | Quy mô nhỏ hơn Grab đáng kể; phải chạy khuyến mãi sâu liên tục để giữ thị phần → biên lợi nhuận mỏng, phụ thuộc vòng gọi vốn để duy trì trợ giá |
| **Xanh SM** | Đội xe điện đồng nhất (VinFast); tài xế là nhân viên, không phải đối tác tự do; giá cố định/mét kiểu taxi truyền thống | Chất lượng dịch vụ đồng đều (xe mới, tài xế được đào tạo chuẩn); hình ảnh sạch/hiện đại; hậu thuẫn vốn mạnh từ Vingroup | Mô hình cực kỳ nặng vốn (mua xe + trả lương tài xế) → mở rộng chậm bằng vốn tự thân, không linh hoạt bằng mô hình nền tảng thuần tuý |
| **Uber** | Người tiên phong surge pricing thuật toán; upfront pricing (hiện giá cố định trước khi đặt); quy mô toàn cầu | Dữ liệu matching/định giá hàng đầu thế giới; upfront pricing minh bạch với rider trước khi xác nhận | Hoa hồng thuộc nhóm cao nhất ngành (nhiều thị trường 25–30%); từng bị chỉ trích đối xử tài xế như nhà thầu thiếu phúc lợi; surge từng gây khủng hoảng truyền thông khi tăng vọt trong tình huống khẩn cấp |
| **Lyft** | Định vị "driver-friendly" hơn Uber tại Mỹ; thương hiệu gần gũi, cộng đồng | Hình ảnh thân thiện với tài xế là USP thật; sẵn sàng thử nghiệm mô hình tip minh bạch | Quy mô nhỏ hơn Uber nhiều lần, khó cạnh tranh vốn; phần lớn phải "theo giá" Uber thay vì dẫn dắt thị trường |
| **DiDi** | Chiến lược trợ giá cực mạnh để chiếm lĩnh thị trường (đã từng đánh bật Uber khỏi Trung Quốc); đa dạng hoá nhanh (xe đạp, giao hàng) | Tốc độ chiếm thị phần cực nhanh; quy mô khổng lồ tạo hiệu ứng mạng lưới mạnh | Mô hình phụ thuộc trợ giá không bền vững nếu thiếu vốn liên tục; từng bị cơ quan quản lý siết chặt vì dữ liệu & độc quyền; giá thấp giả tạo khiến thu nhập tài xế bấp bênh khi trợ giá bị cắt |
| **Bolt** | Hoa hồng thấp hơn Uber (~15–20% tuỳ thị trường) — chiến lược "thách thức giá rẻ có kiểm soát" | Hoa hồng thấp là vũ khí thu hút tài xế hợp lệ (không phải trợ giá không bền vững); mở rộng nhanh vào thị trường ngách (châu Âu, châu Phi) mà Uber/Grab bỏ qua | Thương hiệu yếu hơn ở thị trường lớn; ít tính năng phụ trợ đa dạng; cạnh tranh chủ yếu bằng giá nên biên lợi nhuận mỏng |
| **inDrive** | Mô hình đấu giá ngược độc đáo: rider đề xuất giá, tài xế nhận/từ chối/trả giá; hoa hồng công bố cực thấp, một số thị trường gần như phí cố định thay vì phần trăm | Minh bạch tuyệt đối về hoa hồng với tài xế; tài xế cảm thấy có quyền kiểm soát thật sự; hiệu quả cao ở thị trường mới nổi giá nhạy cảm | Giá cả biến động khó đoán; thuật toán matching/ETA kém hiệu quả hơn mô hình tự động; trải nghiệm "mặc cả" gây khó chịu cho người dùng muốn đặt xe 1 chạm nhanh gọn |
| **Maxim** | Giá cực rẻ, tập trung thành phố nhỏ/thị trường ngách bị các hãng lớn bỏ qua; vận hành tinh gọn, ít tính năng | Hoa hồng thấp, chi phí công nghệ thấp → sống được ở phân khúc thấp mà đối thủ lớn không muốn phục vụ | Thương hiệu gắn với "giá rẻ = chất lượng thấp"; đầu tư an toàn/công nghệ hạn chế; khó nâng cấp trải nghiệm khi thị trường phát triển lên |
| **Yandex Go** | Thuật toán định giá động tinh vi, tích hợp sâu hệ sinh thái Yandex (bản đồ, dữ liệu giao thông riêng) | Hiệu quả vận hành cao nhờ dữ liệu giao thông/bản đồ tự chủ (không phụ thuộc bên thứ ba); tối ưu hoá định giá theo thời gian thực rất mạnh | Phụ thuộc hoàn toàn vào hạ tầng dữ liệu công ty mẹ (khó tái tạo ở thị trường không có hệ sinh thái tương đương); rủi ro địa chính trị/quy định ảnh hưởng thương hiệu toàn cầu |

## 0.3 Panda học được gì — và từ chối gì

| Bài học | Nguồn | Panda áp dụng thế nào |
|---|---|---|
| Hoa hồng thấp là vũ khí thu hút tài xế hợp lệ, không phải "chiêu trò tạm thời" | Bolt | BRB §7.1 đã đặt hoa hồng khởi điểm 20% (Bronze) → 12% (Diamond), thấp hơn mức cao của Uber/Grab. PHẦN 4 tài liệu này còn hạ thấp hơn nữa ở giai đoạn Launch |
| Minh bạch hoa hồng tuyệt đối là lợi thế cạnh tranh thật, không phải rủi ro lộ bí mật | inDrive | BRB §1.2 "Rules Are Public" — Panda công bố hoa hồng công khai cho tài xế, không mập mờ |
| Đầu tư vào dữ liệu bản đồ/giao thông riêng tạo lợi thế dài hạn không phụ thuộc bên thứ ba | Yandex Go | BRB §2.2.2 — Route Engine riêng của Panda đã được thiết kế theo đúng hướng này |
| Bản sắc địa phương là lợi thế cạnh tranh thật với đối thủ ngoại quốc lớn hơn | Be | Panda định vị "hiểu thị trường Việt Nam" thay vì cố cạnh tranh quy mô vốn với Grab |
| Chất lượng dịch vụ đồng nhất tạo niềm tin thương hiệu | Xanh SM | Panda áp dụng qua Driver Tier System + Trust Score (BRB §9.9), không cần sở hữu xe |
| Đa dạng hoá dịch vụ là hướng đi dài hạn đúng | Grab | Đưa vào Roadmap Năm 3 (PHẦN 9), **không** đưa vào giai đoạn hiện tại — tránh dàn trải khi core-product chưa vững |
| **Từ chối:** hoa hồng cao + thuật toán mờ | Grab, Uber | Ngược hoàn toàn với Constitution §1.2 Fairness First |
| **Từ chối:** đốt tiền trợ giá không đo ROI, không có điểm dừng | DiDi | Ngược hoàn toàn với BRB §1.6 Long-Term Sustainability; xem PHẦN 8 |
| **Từ chối:** mô hình đấu giá ngược làm trải nghiệm đặt xe chậm/gây khó chịu | inDrive | Panda giữ trải nghiệm "đặt xe 1 chạm, giá cố định trước khi xác nhận" |
| **Từ chối:** hy sinh an toàn/chất lượng để giữ giá rẻ nhất thị trường | Maxim | Panda không theo đuổi vị trí "rẻ nhất" — xem PHẦN 1 |
| **Từ chối:** mô hình sở hữu tài sản nặng vốn (mua xe, thuê tài xế làm nhân viên) | Xanh SM | Panda là nền tảng đối tác (partner platform), không sở hữu phương tiện |

---

# PHẦN 1 — TRIẾT LÝ ĐỊNH GIÁ

## 1.1 Panda muốn trở thành gì?

Sáu định vị khả dĩ đã được cân nhắc:

| Định vị | Vì sao Panda **không** chọn làm định vị chính |
|---|---|
| **Cheapest** (rẻ nhất) | Mời gọi một cuộc đua xuống đáy với đối thủ có vốn lớn hơn (Xanh SM có Vingroup, Grab có vốn khu vực) — Panda sẽ thua cuộc đua vốn. "Rẻ nhất" cũng mâu thuẫn trực tiếp với thu nhập tài xế bền vững (Ưu tiên #2) |
| **Fairest** (công bằng nhất) | Đây là **một phần** của định vị được chọn, không phải toàn bộ — công bằng mà không minh bạch thì rider/tài xế không cảm nhận được |
| **Most Transparent** (minh bạch nhất) | Cũng là **một phần** — minh bạch mà không công bằng (vd: minh bạch rằng hoa hồng là 30%) thì minh bạch không có giá trị |
| **Driver First** | Đúng về tinh thần (BRB §1.4) nhưng nếu là toàn bộ định vị sẽ mâu thuẫn với thang ưu tiên của chính tài liệu này — Rider đứng ưu tiên #1, trước Driver #2. Driver First là **cách thực thi trong giai đoạn đầu**, không phải đích đến |
| **Passenger First** | Cùng vấn đề ngược lại — không có tài xế thì không có gì để rider dùng. Đây là bài toán con gà quả trứng, không thể chọn một bên làm định vị vĩnh viễn |
| **Balanced** | Đúng nhưng mơ hồ nếu đứng một mình — "cân bằng" giữa cái gì với cái gì cần được định nghĩa rõ |

## 1.2 Định vị được chọn: "Công bằng & Minh bạch — thực thi qua Driver-Enabled Growth"

Panda kết hợp hai trụ cột không thể tách rời và một cơ chế thực thi:

**Trụ cột 1 — Fairest.** Tài xế nhận đúng phần trăm đã công bố, không giảm giữa chừng không báo trước (BRB §1.4: giảm hoa hồng phải báo trước 30 ngày). Rider trả đúng giá đã hiển thị trước khi xác nhận, không phí ẩn (BRB §1.2 Nguyên tắc 1). Hai rider đặt cùng một chuyến, cùng điều kiện, nhận cùng một giá — không định giá cá nhân hoá dựa trên khả năng chi trả cảm nhận được (Constitution Article II §2.1).

**Trụ cột 2 — Most Transparent.** Mọi quy tắc định giá được công bố công khai — ngưỡng surge, bậc hoa hồng, điều kiện khuyến mãi (BRB §1.2 Nguyên tắc 2). Panda có thể giữ bí mật cách **tối ưu hoá** thuật toán (matching, dispatch), nhưng không bao giờ giữ bí mật **quy tắc** quyết định tiền chảy đi đâu.

**Cơ chế thực thi — Driver-Enabled Growth.** Vì thang ưu tiên đặt Rider #1 và Driver #2 liền kề nhau, và vì không có driver thì rider không có gì để dùng, chiến lược thực thi trong 2 năm đầu **nghiêng về phía tài xế trước** một cách có chủ đích (hoa hồng thấp hơn chuẩn dài hạn, guaranteed income mở rộng, xem PHẦN 4). Đây không phải "Driver First" vĩnh viễn — đây là trình tự khởi động hợp lý của một nền tảng hai mặt (two-sided marketplace): **cung đủ tốt trước, thì cầu mới tăng bền vững.**

## 1.3 Điều triết lý này KHÔNG cho phép

- Không giảm giá dưới giá vốn vận hành một cách vô thời hạn để "thắng bằng mọi giá" (xem PHẦN 8).
- Không surge vượt trần 2.0x trong bất kỳ hoàn cảnh nào (đã khoá cứng tại BRB §2.13.3).
- Không định giá khác nhau cho hai rider giống hệt điều kiện chuyến đi dựa trên tín hiệu sẵn sàng chi trả (loại thiết bị, lịch sử thanh toán).
- Không cắt giảm hoa hồng tài xế để tài trợ khuyến mãi rider — mọi khuyến mãi rider do nền tảng tài trợ, tài xế luôn nhận hoa hồng trên giá **trước khuyến mãi** (đã là quy tắc cứng tại BRB §6.5).

---

# PHẦN 2 — ĐỊNH GIÁ CHUYẾN XE

## 2.1 Nguyên tắc thiết kế

Một cấu trúc giá chỉ tốt nếu giải thích được cho một tài xế và một rider trong dưới 60 giây (BRB §2.1). Bảng dưới đây liệt kê từng thành phần được yêu cầu thiết kế, kèm giải thích, và tham chiếu nơi công thức chính xác đã tồn tại (BRB) hoặc đề xuất mới.

## 2.2 Bảng thành phần cước phí

| Thành phần | Thiết kế của Panda | Giải thích |
|---|---|---|
| **Base Fare** | Phí cố định theo thành phố/hạng xe, không nhân surge. Xem BRB §2.2.1 cho bảng giá chính xác (10.000–18.000 VND theo hạng xe) | Bù chi phí tài xế di chuyển đến điểm đón — chi phí này không phụ thuộc quãng đường/thời gian chuyến, nên tách riêng khỏi hai thành phần đó |
| **Included Distance / Included Time** | **Panda chủ động KHÔNG dùng khái niệm này.** Không có "km/phút miễn phí đầu tiên" ẩn trong Base Fare | Một số mô hình taxi truyền thống gộp X km đầu vào "cờ mở cửa" (flag-fall), tạo ra một điểm gián đoạn khó giải thích ("tại sao 2km đầu rẻ hơn km thứ 3?"). Base Fare của Panda đã đóng đúng vai trò "phí mở chuyến" — thêm một lớp "included distance" nữa vi phạm Nguyên tắc 4 của BRB (Simplicity Scales) mà không tạo thêm giá trị |
| **Price / km** | Theo hạng xe, xem BRB §2.2.2 (4.000–5.500 VND/km). Đo bằng Route Engine riêng của Panda, không phụ thuộc Google Maps để tính cước | Tính cước bằng đo lường riêng tránh rủi ro phụ thuộc bên thứ ba thay đổi giá/ngừng dịch vụ — bài học trực tiếp từ Yandex Go (PHẦN 0) |
| **Price / minute** | Theo hạng xe, chỉ kích hoạt khi tốc độ < 10 km/h, xem BRB §2.2.3 (400–550 VND/phút) | Bù cho tài xế khi kẹt xe — quãng đường ngắn nhưng thời gian dài. Loại trừ lẫn nhau với Price/km trong cùng một giây để tránh tính hai lần |
| **Minimum Fare** | Theo hạng xe, xem BRB §2.2.4 (25.000–40.000 VND) | Chuyến rất ngắn vẫn tạo chi phí overhead cố định (matching, xử lý thanh toán, định vị tài xế) mà cước theo mét không đủ bù |
| **Maximum Fare (Price Cap)** | Trần tuyệt đối mỗi chuyến, xem BRB §2.13.6 (500.000 VND nội thành hạng Standard) | Ngăn kịch bản cực đoan (kẹt xe 5 tiếng tạo cước 2 triệu VND) phá vỡ niềm tin dù về mặt kỹ thuật là "đúng công thức" |
| **Airport surcharge** | Phí cố định một lần/chuyến, chia sẻ theo tỷ lệ hoa hồng chuẩn, xem BRB §2.2.7 (10.000 VND) | Bù chi phí vận hành khu vực sân bay cao hơn (chờ, tắc nghẽn) |
| **Late night surcharge** | Hệ số nhân +20% cho chuyến bắt đầu 22:00–05:00, xem BRB §2.2.10 | Lái đêm rủi ro cá nhân cao hơn, giảm khả năng tài xế về nhà đúng giờ |
| **Holiday surcharge** | Hệ số nhân +15%, có thể cộng dồn với Night (trần cộng dồn ×1.50), xem BRB §2.2.11 | Nhu cầu cao dịp lễ trong khi nguồn cung tài xế thường giảm (họ cũng muốn nghỉ lễ) |
| **Rain surcharge** | Hệ số nhân +15%, kích hoạt tự động qua API thời tiết đã xác minh (≥2mm/giờ), xem BRB §2.2.13 | Lái xe trong mưa khó khăn/rủi ro hơn; tự động hoá tránh thiên vị vận hành thủ công |
| **Traffic surcharge** | **Panda không có surcharge kẹt xe riêng biệt.** Đã được xử lý bằng cơ chế Time Fare (kích hoạt khi tốc độ < 10 km/h) | Thêm một surcharge "kẹt xe" riêng sẽ tính hai lần cùng một hiện tượng (đã tính bằng Time Fare) — vi phạm nguyên tắc đơn giản hoá. Route Engine đã tự động phân biệt giờ di chuyển và giờ kẹt xe theo tốc độ GPS thực tế |
| **Bridge fee** — **[MỚI — cần tu chính BRB]** | Pass-through 100% theo chi phí thực tế tài xế khai báo, **không thu hoa hồng**, cùng cơ chế xác thực với Toll Fee (BRB §2.2.8) | Cầu có phí (BOT) hoạt động về bản chất kinh tế giống hệt Toll — không lý do gì để có quy tắc riêng biệt. Đề xuất: gộp Bridge Fee vào cùng nhóm "Toll & Bridge Fee" trong BRB v1.1 |
| **Parking fee** | **[MỚI — cần tu chính BRB]** | Pass-through 100% theo hoá đơn tài xế khai báo khi chờ đón khách tại bãi có phí (vd sân bay, trung tâm thương mại), **không thu hoa hồng**. Yêu cầu ảnh hoá đơn nếu vượt 20.000 VND — chống gian lận, đối chiếu nguyên tắc Fraud-Resistant tại BRB §1.2 Nguyên tắc 5 |
| **Waiting fee** | 500 VND/phút sau 3 phút miễn phí kể từ khi tài xế bấm "Đã đến", xem BRB §2.2.9 | Rider chuẩn bị cần thời gian hợp lý; sau đó thời gian chờ là chi phí thực của tài xế |
| **Cancellation fee** | 10.000 VND (huỷ trước khi tài xế đến, sau 2 phút miễn phí) / 20.000 VND (huỷ sau khi tài xế đã đến) — chia 80/20 tài xế/nền tảng, xem BRB §10.1 | Tài xế đã bỏ thời gian/nhiên liệu di chuyển đến điểm đón — huỷ muộn tạo chi phí thật cần được bù đắp |
| **Long pickup compensation** — **[MỚI — cần tu chính BRB]** | Khi quãng đường tài xế phải di chuyển đến điểm đón (không phải quãng đường chuyến đi) > 3km tại thời điểm ghép chuyến: tài xế nhận thêm 10.000 VND (3–5km) hoặc 20.000 VND (> 5km), **do nền tảng chi trả toàn bộ, rider không phải trả thêm một đồng nào** | Đây là một lựa chọn thiết kế khác biệt có chủ đích: việc ghép một tài xế ở xa là **lỗi hiệu quả của thuật toán dispatch**, không phải lỗi của rider — bắt rider trả thêm tiền vì thuật toán ghép chuyến chưa tối ưu vi phạm Constitution Article II §2.1 (Rider Fairness). Đồng thời khoản bù này giảm động lực tài xế từ chối các chuyến đón xa, cải thiện Acceptance Rate (BRB §9.1) mà không cần phạt rider |

## 2.3 Toàn bộ ví dụ tính cước chi tiết

Xem BRB §2.17 và §2.17B — ba ví dụ đầy đủ (chuyến ngắn không phụ phí, chuyến sân bay giờ cao điểm, chuyến đêm giao thừa mưa surge ×1.6) đã minh hoạ cách các thành phần trên kết hợp theo đúng thứ tự tính toán. Tài liệu này không lặp lại các ví dụ đó.

---

# PHẦN 3 — DRIVER REVENUE

## 3.1 Driver nhận bao nhiêu?

Hệ thống hoa hồng theo tier đã được thiết kế đầy đủ tại BRB §7.1: **80% (Bronze) → 82% (Silver) → 84% (Gold) → 86% (Platinum) → 88% (Diamond)**. Đây là mức **thấp hơn** nhóm cao của Uber/Grab (thường 70–75% cho tài xế) và tương đương/tốt hơn Bolt — một lựa chọn chiến lược có chủ đích, không phải ngẫu nhiên (xem PHẦN 0.3).

## 3.2 Commission bao nhiêu?

Phần bù của bảng trên: **20% → 12%** tuỳ tier. Áp dụng trên (Base + Distance + Time + surcharge), **không** áp dụng trên Booking Fee (100% về nền tảng) và Toll/Bridge/Parking (100% pass-through, 0% hoa hồng) — BRB §2.2.6.

## 3.3 Platform Fee / Service Fee

Panda không có "Platform Fee" tách biệt khỏi Commission — gộp làm một để tránh rider/tài xế nhìn thấy hai dòng phí trông giống nhau nhưng đi vào hai công thức khác nhau (vi phạm nguyên tắc "giải thích trong 60 giây" — PHẦN 2.1). **Booking Fee** (2.000 VND/chuyến, BRB §2.2.5) đóng vai trò "service fee" — cố định, không surge, 100% về nền tảng, hiển thị minh bạch thành dòng riêng.

## 3.4 Insurance

Chưa có sản phẩm bảo hiểm được mua (BRB Appendix B, UBD-004 — "Insurance Partnership" **chưa có quyết định**, cần Legal + CFO phê duyệt trước khi ra mắt thương mại). Khuyến nghị chiến lược của tài liệu này: bảo hiểm tai nạn chuyến đi cơ bản nên là **chi phí nền tảng gánh** (không cộng vào cước), coi như chi phí vận hành bắt buộc — tương tự cách Grab/Uber xử lý bảo hiểm cơ bản ở hầu hết thị trường, để không làm giá Panda kém cạnh tranh vì một dòng phí bảo hiểm hiển thị riêng.

## 3.5 VAT

Nằm ngoài phạm vi tài liệu giá — theo BRB §6.10, thuế nền tảng (VAT, thuế doanh nghiệp) do bộ phận Tài chính tính toán và nộp độc lập, tài xế tự chịu trách nhiệm thuế thu nhập cá nhân (được phân loại đối tác độc lập). Tài liệu này không thay đổi phân loại này.

## 3.6 Thu hộ (Toll/Bridge/Parking pass-through)

100% về tài xế, 0% hoa hồng — xem PHẦN 2.2. Đây là tiền tài xế đã tạm ứng thay rider, không phải doanh thu của ai cả.

## 3.7 Tiền tip

**Chưa được hỗ trợ** (BRB Appendix B, UBD-005 — "Deferred. Until a decision is made, tipping is not supported"). **Khuyến nghị chiến lược:** tài liệu này đề xuất chính thức mở khoá tipping tuỳ chọn từ Giai đoạn 3 (2.000 tài xế, PHẦN 4) vì: (a) chi phí triển khai gần như bằng 0 với nền tảng (100% về tài xế, không qua hoa hồng), (b) tăng thu nhập tài xế thuần tuý — đúng hướng thang ưu tiên #2, (c) không nền tảng lớn nào trong PHẦN 0 làm tốt điều này — đây là khoảng trống khác biệt hoá thật sự. Đây là **đề xuất**, cần quyết định CPO chính thức theo đúng quy trình đóng UBD-005.

## 3.8 Thưởng (Incentive)

Toàn bộ hệ thống Quest/Streak/Peak/Airport/Rain/Guaranteed Income/Acceptance/Completion/Rating/Referral Bonus đã thiết kế đầy đủ tại BRB Part 8 — xem PHẦN 6 tài liệu này cho cách các thưởng này được **kích hoạt theo giai đoạn tăng trưởng**, không phải bật tất cả cùng lúc từ ngày đầu.

## 3.9 Phạt

Cơ chế clawback khi gian lận (BRB §8.14 Rule 4, §11 Fraud Rules) và giảm ưu tiên dispatch khi Cancellation Rate cao (BRB §9.3) đã đầy đủ. Tài liệu này không thêm hình phạt mới — nguyên tắc BRB §1.4 "tài khoản tài xế không bao giờ bị đình chỉ mà không có lý do bằng văn bản và một lộ trình khiếu nại" là bất biến.

## 3.10 Hoàn tiền

Khi refund do lỗi nền tảng/tài xế, tài xế **giữ nguyên** thu nhập nếu lỗi thuộc về nền tảng, hoặc bị thu hồi nếu lỗi thuộc về tài xế (BRB §6.7) — nguyên tắc "driver luôn được tính trên giá đầy đủ trước khuyến mãi" không bao giờ bị vi phạm bởi refund.

---

# PHẦN 4 — GIAI ĐOẠN PHÁT TRIỂN

Đây là phần **chiến lược mới hoàn toàn**, không có trong BRB — BRB định nghĩa hoa hồng theo **hiệu suất từng tài xế** (tier cá nhân), còn phần này định nghĩa chiến lược giá theo **quy mô toàn nền tảng** (bao nhiêu tài xế/rider đã tham gia). Hai hệ thống này chạy song song và không mâu thuẫn: một tài xế Bronze ở Giai đoạn 0 vẫn nhận ưu đãi hoa hồng của Giai đoạn 0, chồng lên trên biểu tier cá nhân khi tier đó được kích hoạt đầy đủ.

## 4.1 Giai đoạn 0 — Launch (1 thành phố, 0–99 tài xế)

**Mục tiêu:** chứng minh mô hình vận hành được. Đây là giai đoạn sống còn (existential) — không phải giai đoạn tăng trưởng.

| Đòn bẩy | Chiến lược |
|---|---|
| Commission | **10% flat cho mọi tài xế** — thấp hơn cả mức Bronze chuẩn (20%) trong BRB. Đây là trợ giá hoa hồng tạm thời, có hạn, để thuyết phục từng tài xế đầu tiên (ở quy mô này, mỗi tài xế là một quyết định cá nhân quan trọng) |
| Khuyến mãi | First Ride Promotion nâng lên 70% off (so với chuẩn 50%/30.000 VND cap tại BRB §3.2.1) — ngân sách tách riêng "Launch Budget", không tính vào ngân sách campaign thường xuyên |
| Thưởng | Guaranteed Income Programme (BRB §8.9) **mở rộng không giới hạn 90 ngày** — áp dụng cho mọi tài xế cho đến khi nền tảng đạt mốc 100 tài xế, không chỉ 90 ngày đầu của riêng từng tài xế |
| Chiến lược giá | Giá ngang hoặc thấp hơn taxi truyền thống một chút tại khu vực đó để xoá rào cản dùng thử lần đầu. **Không** cần thấp hơn Grab/Be — ở quy mô này Panda chưa đối đầu trực tiếp, mục tiêu là **có mặt**, không phải **thắng** |

## 4.2 Giai đoạn 1 — 100 tài xế

**Mục tiêu:** đạt mật độ tối thiểu để ETA đủ tốt (rider không chờ quá 10 phút).

| Đòn bẩy | Chiến lược |
|---|---|
| Commission | Tăng dần về **15%** — vẫn thấp hơn Bronze chuẩn nhưng bắt đầu thu hẹp khoảng cách với biểu dài hạn |
| Khuyến mãi | First Ride trở về đúng chuẩn BRB (50%/30.000 VND); kích hoạt Golden Hour Promotion (BRB §3.2.3) để định hình thói quen dùng ngoài giờ cao điểm |
| Thưởng | Kích hoạt đầy đủ Daily Quest + Weekly Mission (BRB §8.2–8.4) để giữ đủ tài xế online tạo ETA tốt |
| Chiến lược giá | Bắt đầu theo dõi giá Grab/Be tại khu vực; neo giá thấp hơn 5–10% để tạo lý do dùng thử **rõ ràng**, không chỉ ngang bằng |

## 4.3 Giai đoạn 2 — 500 tài xế

**Mục tiêu:** đủ mật độ để Dynamic Surge có ý nghĩa thống kê (cần đủ dữ liệu DSR theo từng zone — xem PHẦN 5).

| Đòn bẩy | Chiến lược |
|---|---|
| Commission | Đạt đúng biểu chuẩn BRB (Bronze 20%, tier cao hơn theo hiệu suất cá nhân) — trợ giá Launch chính thức kết thúc |
| Khuyến mãi | Giảm cường độ subsidy trên mỗi chuyến; chuyển ngân sách sang Referral Programme (BRB §3.2.7) — tăng trưởng lan truyền rẻ hơn quảng cáo trả phí |
| Thưởng | Kích hoạt Peak/Airport/Rain Bonus (BRB §8.6–8.8) — đủ tài xế để có nhu cầu rõ rệt theo khu vực/thời điểm |
| Chiến lược giá | Cho phép Dynamic Surge hoạt động thật (trần ×2.0 theo BRB §2.13.3) — mốc đầu tiên Panda có công cụ cân bằng cung–cầu thực sự thay vì chỉ giảm giá |

## 4.4 Giai đoạn 3 — 2.000 tài xế

**Mục tiêu:** mở rộng đa thành phố, bắt đầu cạnh tranh trực diện.

| Đòn bẩy | Chiến lược |
|---|---|
| Commission | Tier system đầy đủ vận hành (Bronze→Diamond); trung bình toàn nền tảng ước tính ~17–18% (nhiều tài xế đã lên Silver/Gold) |
| Khuyến mãi | Pilot Membership/Subscription (định hướng BRB §15.5) — bắt đầu khoá người dùng trung thành bằng lợi ích thay vì giảm giá lặp lại |
| Thưởng | Đẩy mạnh Driver Referral Bonus (BRB §8.13) — tài xế hiện tại là kênh tuyển dụng rẻ nhất; mở khoá tipping (PHẦN 3.7) |
| Chiến lược giá | Sẵn sàng Competitive Response Protocol có kiểm soát (PHẦN 8) — đây là giai đoạn đối thủ lớn bắt đầu chú ý tới Panda |

## 4.5 Giai đoạn 4 — 10.000 tài xế

**Mục tiêu:** hoà vốn vận hành (break-even) ở các thành phố đã ra mắt trên 12 tháng.

| Đòn bẩy | Chiến lược |
|---|---|
| Commission | Take Rate toàn nền tảng chạm biên dưới mục tiêu BRB §14.6 (12–16%) |
| Khuyến mãi | Chuyển từ "mua tăng trưởng" sang "giữ chân có ROI" — mọi campaign phải đạt CPIR mục tiêu (BRB §3.8) nghiêm ngặt hơn giai đoạn trước |
| Thưởng | Chuẩn hoá toàn bộ Driver Incentive Engine đúng theo BRB Part 8 — không còn ưu đãi "giai đoạn khởi động" nào |
| Chiến lược giá | Bắt đầu thử nghiệm hạng xe giá trị gia tăng (Premium, XL — đã có trong BRB §2.2.1) để tăng ARPU thay vì giảm giá hạng Standard |

## 4.6 Giai đoạn 5 — 100.000 người dùng (mốc rider, chạy song song với mốc tài xế)

**Mục tiêu:** lợi nhuận bắt đầu xuất hiện có kiểm soát — **không đánh đổi công bằng tài xế để đạt được.**

| Đòn bẩy | Chiến lược |
|---|---|
| Commission | Có thể tinh chỉnh nhẹ lên nếu Take Rate dưới mục tiêu, **luôn công bố trước 30 ngày** (bất biến theo BRB §1.4) |
| Khuyến mãi | Cá nhân hoá theo hành vi thay vì giảm giá đại trà — vd Win-back Campaign cho rider ngừng hoạt động > 30 ngày (PHẦN 7) |
| Thưởng | Tier cao (Platinum/Diamond) nhận thêm quyền lợi phi tiền mặt (hỗ trợ pháp lý, ưu tiên bảo hiểm) để giữ chân mà không cần tăng chi tiền mặt liên tục |
| Chiến lược giá | Điểm chuyển từ "công cụ tăng trưởng" sang "công cụ tối ưu bền vững" — lợi nhuận lần đầu tiên được cân nhắc chính thức, đúng như thang ưu tiên đã định ở đầu tài liệu |

---

# PHẦN 5 — DYNAMIC PRICING

## 5.1 Nguyên tắc: Logic quyết định, không phải AI

Toàn bộ hệ thống dưới đây là **rule engine tất định** (deterministic) — mỗi tín hiệu đầu vào ánh xạ sang một hệ số/phụ phí cụ thể theo bảng tra cứu công khai, không có mô hình học máy "hộp đen" nào quyết định giá. Điều này khớp trực tiếp với BRB §1.2 Nguyên tắc 2 ("Rules Are Public") — một rider có thể tự tính ra vì sao giá của họ là con số đó.

## 5.2 Các tín hiệu đầu vào và cách xử lý

| Tín hiệu | Cách xử lý | Cơ chế |
|---|---|---|
| **Demand** | Số yêu cầu đặt xe đang hoạt động trong 5 phút gần nhất theo zone | Đầu vào của DSR — xem 5.3 |
| **Supply** | Số tài xế sẵn sàng trong/gần zone | Đầu vào của DSR — xem 5.3 |
| **Surge** | Hệ số nhân theo bảng DSR, trần ×2.0 | Đã định nghĩa đầy đủ tại BRB §2.13 — tài liệu này không lặp lại bảng DSR, chỉ tham chiếu |
| **Heat Map** | Lưới zone địa lý, tô màu theo DSR hiện tại (xanh = bình thường, vàng = bận, cam = cao, đỏ = rất cao) | Hiển thị cho **cả hai phía**: rider thấy "khu vực đang đông, giá có thể cao hơn" trước khi mở app đặt xe; tài xế thấy gợi ý "di chuyển đến khu vực đỏ để có nhiều chuyến hơn" — thuần tuý tổng hợp + ngưỡng màu, không dự đoán |
| **Weather** | Rain Surcharge tự động qua API thời tiết đã xác minh (≥2mm/giờ), xem BRB §2.2.13. Mở rộng: điều kiện thời tiết cực đoan khác (bão, ngập) kích hoạt cùng cơ chế nhưng do Operations bật thủ công (không có API tin cậy cho "ngập lụt cục bộ") | Ngưỡng cố định + xác nhận thủ công cho trường hợp API không phủ được |
| **Traffic** | Không phải một surcharge riêng — xử lý qua Time Fare (PHẦN 2.2) + Route Engine hiển thị ETA thực tế | Tách biệt rõ với Surge: traffic ảnh hưởng **thời gian** chuyến (Time Fare), Surge ảnh hưởng **hệ số giá** do mất cân bằng cung–cầu. Hai khái niệm khác nhau, không gộp |
| **Festival** | Holiday Surcharge (+15%, BRB §2.2.11) + Festival Promotion đồng thời kích hoạt (BRB §3.2.6) — phụ phí bù tài xế, khuyến mãi giữ rider không thấy giá "nhảy vọt" dịp lễ | Danh sách ngày lễ quản lý trong admin portal theo thành phố/quốc gia |
| **Airport** | Airport Fee cố định (BRB §2.2.7) + Airport queue priority cho tier cao (BRB §7.5) | Polygon địa lý cố định định nghĩa trong admin portal cho từng sân bay |
| **Concert / sự kiện lớn** — **[MỚI — cần tu chính BRB]** | "Event Zone": geofence tạm thời quanh địa điểm (sân vận động, trung tâm hội nghị) trong khung giờ sự kiện đã biết trước (lịch sự kiện nhập thủ công bởi Operations), kích hoạt phụ phí sự kiện tương tự Airport Fee nhưng có hiệu lực **tạm thời và có thời hạn rõ ràng** | Cùng cơ chế Airport Fee nhưng geofence động theo lịch sự kiện thay vì cố định vĩnh viễn. Đề xuất mức khởi điểm: bằng Airport Fee (10.000 VND), điều chỉnh theo dữ liệu thực tế sau khi vận hành thử |
| **Night** | Night Surcharge +20%, 22:00–05:00 (BRB §2.2.10) | Đã đầy đủ, không thay đổi |

## 5.3 Thứ tự áp dụng khi nhiều tín hiệu trùng nhau

Không thay đổi thứ tự đã xác lập tại BRB §2.17 (Surge áp dụng trước trên sub-total gốc, sau đó Night/Holiday/Rain nhân tuần tự, trần cộng dồn ×1.60) và Appendix A Issue 1. Event Zone Surcharge (mới) được xử lý **cùng nhóm với Airport Fee** — cộng dồn cố định, không nhân với Surge.

---

# PHẦN 6 — DRIVER INCENTIVE

## 6.1 Nguyên tắc

BRB §8.1 đã xác lập rõ: incentive tồn tại để (1) tăng nguồn cung giờ cao điểm, (2) cải thiện chất lượng, (3) giữ chân tài xế giỏi, (4) tạo thu nhập phụ **dự đoán được**. Incentive **không bao giờ** được thiết kế để ép tài xế làm việc quá sức hoặc khuyến khích hành vi hại rider (nhận chuyến rồi huỷ). Tài liệu này không thay đổi nguyên tắc này — chỉ định nghĩa **khi nào từng cơ chế được bật**, gắn với giai đoạn tăng trưởng ở PHẦN 4.

## 6.2 Toàn bộ danh mục (tham chiếu BRB Part 8) và thời điểm kích hoạt

| Cơ chế | Định nghĩa đầy đủ | Kích hoạt từ giai đoạn |
|---|---|---|
| Daily Quest | BRB §8.2 | Giai đoạn 1 (100 tài xế) |
| Weekly Mission | BRB §8.3 | Giai đoạn 1 |
| Monthly Incentive | BRB §8.4 | Giai đoạn 2 (500 tài xế) |
| Streak Bonus | BRB §8.5 | Giai đoạn 1 |
| Peak Hour Bonus | BRB §8.6 | Giai đoạn 2 |
| Airport Bonus | BRB §8.7 | Giai đoạn 2 |
| Rain Bonus | BRB §8.8 | Giai đoạn 0 (ngay từ Launch — chi phí thấp, tác động giữ chân cao) |
| Guaranteed Income Programme | BRB §8.9 | Giai đoạn 0, dạng mở rộng (PHẦN 4.1); trở về chuẩn 90-ngày/tài-xế từ Giai đoạn 2 |
| Acceptance Bonus | BRB §8.10 | Giai đoạn 2 |
| Completion Bonus | BRB §8.11 | Giai đoạn 2 |
| Rating Bonus | BRB §8.12 | Giai đoạn 3 (2.000 tài xế — cần đủ volume đánh giá để công bằng) |
| Referral Bonus (tài xế) | BRB §8.13 | Giai đoạn 1, đẩy mạnh ngân sách từ Giai đoạn 3 |

## 6.3 Vì sao không bật tất cả từ ngày đầu

Ở Giai đoạn 0 (dưới 100 tài xế), hầu hết các bonus theo hiệu suất tương đối (Rating Bonus, Acceptance Bonus) không có ý nghĩa thống kê — mẫu quá nhỏ để so sánh công bằng. Bật quá sớm tạo cảm giác bất công (một tài xế may mắn có vài chuyến tốt "leo top" không phản ánh chất lượng thật). Nguyên tắc: **incentive theo phần trăm/so sánh** chỉ bật khi đủ khối lượng dữ liệu; **incentive tuyệt đối** (Guaranteed Income, Rain Bonus) có thể bật ngay vì không cần so sánh giữa các tài xế.

---

# PHẦN 7 — PASSENGER PROMOTION

## 7.1 Danh mục đã có trong BRB (tham chiếu, không lặp lại công thức)

| Loại | Định nghĩa đầy đủ |
|---|---|
| First Ride | BRB §3.2.1 |
| Referral (rider) | BRB §3.2.7 |
| Golden Hour | BRB §3.2.3 |
| Weekend | BRB §3.2.4 |
| Rain Campaign | BRB §3.2.5 |
| Festival | BRB §3.2.6 |
| Birthday | BRB §3.2.2 |
| Cashback | BRB §3.2.8 |
| Coupon Campaign | BRB §3.2.9 |

## 7.2 Thiết kế mới cho các mục người dùng yêu cầu chưa có công thức đầy đủ trong BRB

### 7.2.1 Membership / Subscription — **[MỚI — cần tu chính BRB, đã có định hướng tại BRB §15.5]**

**Tên đề xuất:** Panda Plus. **Cơ chế:** phí thuê bao hàng tháng cố định, đổi lại: booking fee giảm (hoặc miễn), trần surge cá nhân thấp hơn trần chung (vd ×1.5 thay vì ×2.0 cho thành viên), ưu tiên matching nhẹ (tương tự priority dispatch của driver tier, BRB §7.5, nhưng áp cho rider). **Thời điểm ra mắt:** pilot Giai đoạn 3 (PHẦN 4.4) — cần đủ volume chuyến để phí thuê bao có giá trị cảm nhận rõ ràng.

### 7.2.2 Student — **[MỚI — cần tu chính BRB]**

**Mục tiêu:** hình thành thói quen sử dụng sớm ở nhóm có LTV dài hạn cao (sinh viên → nhân viên văn phòng trong vài năm tới), chấp nhận biên lợi nhuận thấp ở giai đoạn này. **Xác minh:** email trường học (.edu.vn hoặc danh sách domain đã đối tác) hoặc thẻ sinh viên tải lên thủ công, duyệt bởi Operations. **Ưu đãi:** 10% off, tối đa 5 chuyến/tuần, không giới hạn thời gian (khác First Ride — đây là ưu đãi dài hạn theo phân khúc, không phải one-time). **Chống lạm dụng:** một email trường học chỉ xác minh một tài khoản, tái xác minh mỗi năm học.

### 7.2.3 Corporate — tham chiếu định hướng BRB §15.4

Không thiết kế lại — BRB đã định hướng: tài khoản công ty có hạn mức tín dụng (post-pay), quy trình duyệt chuyến, tích hợp báo cáo chi phí. Tài liệu này chỉ xác nhận: Corporate **không** nằm trong 5 giai đoạn tăng trưởng cốt lõi ở PHẦN 4 — đây là hướng mở rộng doanh thu độc lập, cân nhắc từ Năm 2 của Roadmap (PHẦN 9).

### 7.2.4 Comeback / Win-back — **[MỚI — cần tu chính BRB]**

**Đối tượng:** rider không có chuyến nào trong 30 ngày liên tiếp, từng có ít nhất 3 chuyến trước đó (loại trừ tài khoản mới/gian lận vốn nên nhận First Ride thay vì Win-back). **Ưu đãi:** 30% off, tối đa 25.000 VND, hiệu lực 7 ngày kể từ khi được gửi. **Đo lường:** đúng khung ROI của BRB §3.8 (CPIR, retention uplift D30) — nếu một chiến dịch Win-back không đạt ngưỡng ROI sau khi thử nghiệm, dừng lại, không lặp lại theo quán tính.

### 7.2.5 Birthday — đã có đầy đủ tại BRB §3.2.2, không cần thiết kế thêm.

---

# PHẦN 8 — CHIẾN LƯỢC CẠNH TRANH

## 8.1 Nguyên tắc chung: "Competitive Response Protocol"

Panda **không phản ứng tự động** theo từng động thái giá của đối thủ. Mọi phản ứng phải đi qua một quy trình có kiểm soát:

1. **Đo lường trước khi phản ứng.** Theo dõi churn thực tế (rider/driver rời bỏ theo tuần) trong 7–14 ngày sau khi đối thủ thay đổi giá — không phản ứng dựa trên tin đồn hoặc lo sợ.
2. **Ưu tiên phản ứng bằng giá trị trước khi phản ứng bằng giá.** Tăng cường minh bạch, nhấn mạnh không phí ẩn, cải thiện chất lượng dịch vụ — trước khi cân nhắc giảm giá.
3. **Nếu phải phản ứng bằng giá, luôn có mục tiêu (targeted), không đại trà.** Voucher gửi đúng nhóm rider có nguy cơ rời bỏ thật sự (dựa trên usage pattern), không giảm giá toàn bộ nền tảng.
4. **Mọi phản ứng vượt ngân sách khuyến mãi hiện tại phải qua CFO + CPO phê duyệt** — đúng cơ chế Campaign Budget Rules tại BRB §3.3, không có ngoại lệ "tình huống khẩn cấp cạnh tranh".
5. **Mọi phản ứng có điểm kết thúc rõ ràng** (thời gian + ngân sách + KPI mục tiêu) được xác định **trước khi** bắt đầu — không có "cuộc chiến giá mở" (open-ended war).

## 8.2 Kịch bản cụ thể

**Nếu Grab giảm giá mạnh:** Không giảm giá theo ngay lập tức. Grab có biên lợi nhuận từ hệ sinh thái đa dịch vụ (ads, fintech) để trợ giá ride trong thời gian dài — Panda không có nguồn lực đó và sẽ thua nếu chạy đua trực tiếp. Phản ứng: kích hoạt voucher có mục tiêu cho nhóm rider rủi ro rời bỏ cao nhất, đồng thời truyền thông rõ "giá Panda không có phí ẩn" (nhiều rider phàn nàn về surge/phí Grab không rõ ràng — đây là điểm khác biệt thật).

**Nếu Be miễn phí (chạy campaign 0đ):** Đây thường là chiến dịch ngắn hạn có ngân sách giới hạn, không bền vững — Panda theo dõi, **không chạy đua "ai lỗ nhiều hơn."** Tận dụng cơ hội: trong đợt miễn phí, tài xế Be thường bị quá tải hoặc thu nhập bị pha loãng — đây là thời điểm tốt để đẩy mạnh Driver Referral Bonus (PHẦN 6) nhắm đến tài xế đang bất mãn.

**Nếu Xanh SM trợ giá:** Xanh SM cạnh tranh bằng chất lượng đồng nhất/EV, không phải giá rẻ nhất (mô hình nặng vốn của họ không cho phép giảm giá vô thời hạn — PHẦN 0). Panda không đối đầu trực diện phân khúc "xe điện cao cấp đồng nhất" — tập trung vào phân khúc giá hợp lý + mạng lưới đối tác linh hoạt, điều mà mô hình nhân viên của Xanh SM không có được.

## 8.3 "Không đốt tiền vô tội vạ" — vận hành hoá

- Mọi ngân sách phản ứng cạnh tranh nằm trong cùng khuôn khổ Campaign Budget Rules đã có (BRB §3.3) — không có "quỹ chiến tranh" riêng nằm ngoài kiểm soát tài chính thông thường.
- Vũ khí cạnh tranh chính của Panda **không phải giá thấp nhất** — là: (1) driver trust (thanh toán đúng hạn, hoa hồng minh bạch, không giảm đột ngột), (2) rider trust (không phí ẩn, giá hiển thị = giá trả), (3) Route Engine riêng không phụ thuộc bên thứ ba, (4) bản sắc "hiểu thị trường Việt Nam".
- Đây chính là lý do PHẦN 1 chọn "Fairest & Most Transparent" thay vì "Cheapest" — một cuộc chiến giá sẽ luôn thua trước đối thủ có vốn lớn hơn; một cuộc chiến niềm tin thì không.

---

# PHẦN 9 — ROADMAP 3 NĂM

| | **Năm 1** | **Năm 2** | **Năm 3** |
|---|---|---|---|
| **Quy mô tài xế** | Launch → 2.000 (Giai đoạn 0–3) | 2.000 → 10.000 (Giai đoạn 4) | > 10.000, đa thành phố |
| **Quy mô rider** | Xây dựng tại 1–2 thành phố | Chạm mốc 100.000 (Giai đoạn 5) | Tăng trưởng hữu cơ + referral chiếm ưu thế |
| **Commission trung bình** | ~15–18% (nhiều tài xế vẫn ở tier thấp/đang hưởng ưu đãi Launch) | Tiến về mục tiêu Take Rate 12–16% (BRB §14.6) | Ổn định tại 12–16%, có thể tinh chỉnh theo dữ liệu thực tế (luôn báo trước 30 ngày) |
| **Mức lợi nhuận** | **Âm theo kế hoạch** — đầu tư tăng trưởng có kiểm soát, không phải lỗ mất kiểm soát. Mọi khoản lỗ đều gắn với ngân sách campaign đã duyệt (BRB §3.3), không có chi tiêu ngoài kế hoạch | **Hoà vốn** ở thành phố ra mắt đầu tiên (Giai đoạn 4); vẫn đầu tư tăng trưởng ở thành phố mới mở | **Dương có kiểm soát** ở các thành phố trưởng thành; tái đầu tư lợi nhuận vào thành phố mới thay vì rút ra ngay |
| **Mục tiêu chính** | Chứng minh mô hình vận hành được; đạt density đủ để Surge có ý nghĩa; network effect hình thành ở 1–2 thành phố | Mở rộng đa thành phố; Membership pilot; Driver Tier System trưởng thành đầy đủ; tipping ra mắt | Đa dạng hoá doanh thu (EV programme, Corporate, cân nhắc pilot Food Delivery theo BRB §15.1); network effect đủ mạnh để giảm phụ thuộc khuyến mãi liên tục |

**Nguyên tắc xuyên suốt cả 3 năm:** không năm nào lợi nhuận được ưu tiên cao hơn 5 mục tiêu ở đầu tài liệu. Năm 3 lợi nhuận dương **là hệ quả** của tăng trưởng đúng trình tự, không phải mục tiêu được theo đuổi trực tiếp bằng cách cắt giảm chi phí tài xế hoặc tăng giá rider đột ngột.

---

# PHẦN 10 — KẾT LUẬN

## 10.1 Mô hình giá cuối cùng

Panda vận hành một **mô hình cước theo thành phần** (component-based fare, đã định nghĩa đầy đủ tại BRB Part 2 và PHẦN 2 tài liệu này), với:

- **Triết lý:** Fairest & Most Transparent, thực thi qua trình tự Driver-Enabled Growth (PHẦN 1).
- **Cấu trúc phí:** Base + Distance + Time + Surcharge có kiểm soát (trần cộng dồn ×1.60 tĩnh, Surge riêng trần ×2.0), không có phí ẩn, không "included distance" gây khó hiểu.
- **Hoa hồng:** 20% → 12% theo tier hiệu suất cá nhân (BRB §7.1), cộng thêm ưu đãi tạm thời sâu hơn theo giai đoạn tăng trưởng nền tảng (10% → 15% → chuẩn, PHẦN 4) — hai lớp độc lập, không mâu thuẫn.
- **Tăng trưởng:** 6 giai đoạn rõ ràng từ Launch đến 100.000 người dùng, mỗi giai đoạn có công cụ (commission/khuyến mãi/thưởng/chiến lược giá) khác nhau, được thiết kế để không bao giờ tối ưu một chỉ số bằng cách hy sinh chỉ số ưu tiên cao hơn nó.
- **Cạnh tranh:** không có "chiến tranh giá" — cạnh tranh bằng niềm tin, minh bạch, và một Route Engine không phụ thuộc bên thứ ba.

## 10.2 Vì sao mô hình này thắng về dài hạn

Ba nguyên tắc bất biến, không bao giờ được sửa đổi dù dưới áp lực cạnh tranh hay áp lực tăng trưởng nào:

1. **Surge không bao giờ vượt ×2.0.** (BRB §2.13.3 — bất biến)
2. **Giá hiển thị trước khi xác nhận = giá rider trả.** (Constitution §1.2, BRB §1.2 Nguyên tắc 1 — bất biến)
3. **Tài xế luôn được tính hoa hồng trên giá đầy đủ trước khuyến mãi, không bao giờ gánh chi phí khuyến mãi do nền tảng quyết định.** (BRB §6.5 — bất biến)

Các đối thủ trong PHẦN 0 mà Panda nghiên cứu đều thắng nhanh ở đâu đó bằng cách phá vỡ một trong ba nguyên tắc này (Grab/Uber: minh bạch; DiDi: kỷ luật chi tiêu; Xanh SM: linh hoạt mô hình). Panda chấp nhận tăng trưởng **chậm hơn** trong Năm 1–2 để giữ nguyên cả ba nguyên tắc — và đặt cược rằng một nền tảng giữ được niềm tin của cả rider lẫn driver trong 3 năm đầu sẽ còn đứng vững khi các đối thủ dựa vào trợ giá phải điều chỉnh giá lên (như lịch sử Grab, Uber, DiDi đã từng trải qua sau giai đoạn "đốt tiền" ban đầu).

**Lợi nhuận sẽ đến sau — đúng như mục tiêu đã đặt ra ở đầu tài liệu.**

---

## PHỤ LỤC A — Danh sách quy tắc mới cần tu chính vào Business Rule Bible

Theo đúng quy trình quản trị tài liệu đã thiết lập (Constitution Article XI — Amendment Process), các quy tắc sau được **đề xuất** trong tài liệu chiến lược này nhưng **chưa có hiệu lực vận hành** cho đến khi được tu chính chính thức vào BRB (dự kiến v1.1):

| # | Quy tắc mới | Vị trí đề xuất trong BRB | Tham chiếu PHẦN |
|---|---|---|---|
| 1 | Bridge Fee (gộp cùng Toll Fee) | §2.2.8 mở rộng | 2.2 |
| 2 | Parking Fee | §2.2 thêm mục mới | 2.2 |
| 3 | Long Pickup Compensation | §2.2 thêm mục mới | 2.2 |
| 4 | Event Zone Surcharge (concert/sự kiện) | §2.2 thêm mục mới, cạnh Airport Fee | 5.2 |
| 5 | Mở khoá Tipping (đóng UBD-005) | Appendix B → Part mới "Tipping" | 3.7 |
| 6 | Panda Plus Membership/Subscription (cụ thể hoá §15.5) | §15.5 mở rộng thành Part đầy đủ | 7.2.1 |
| 7 | Student Promotion | Part 3 thêm mục 3.2.10 | 7.2.2 |
| 8 | Comeback/Win-back Promotion | Part 3 thêm mục 3.2.11 | 7.2.4 |
| 9 | Bảng hoa hồng theo giai đoạn tăng trưởng nền tảng (0–5) | Part 7 mở rộng, tách biệt với tier cá nhân §7.1 | 4 |

## PHỤ LỤC B — Câu hỏi mở cần quyết định trước khi triển khai

Kế thừa từ BRB Appendix B, các mục sau **ảnh hưởng trực tiếp** đến chiến lược giá trong tài liệu này và cần được đóng trước khi Giai đoạn tương ứng bắt đầu:

- **UBD-001 (Cash Payment):** ảnh hưởng trực tiếp đến khả năng triển khai ở thị trường yêu cầu tiền mặt — cần đóng trước Giai đoạn 1.
- **UBD-005 (Tipping):** cần đóng chính thức trước Giai đoạn 3 (PHẦN 3.7).
- **UBD-004 (Insurance Partnership):** cần đóng trước khi ra mắt thương mại chính thức (ảnh hưởng PHẦN 3.4).
- **UBD-006 (Scheduled Rides pricing lock):** cần đóng trước khi tính năng đặt lịch được xây dựng — không nằm trong phạm vi 3 năm đầu của roadmap này nhưng cần lưu ý cho Năm 3.

---

*Kết thúc tài liệu — Panda Master Pricing Strategy — v0.1 (Draft)*
*Tài liệu này phải được CEO, CPO, CFO phê duyệt trước khi bất kỳ quy tắc **[MỚI]** nào trong PHỤ LỤC A được triển khai.*
*Không có dòng code, API, hay thay đổi backend nào được tạo ra từ tài liệu này — đây thuần tuý là tài liệu chiến lược kinh doanh.*
