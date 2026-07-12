# Panda — Delivery V1: Thiết kế Chức năng Giao hàng (Design)

**Document Classification:** Internal — Confidential
**Authority:** CPO · COO · CTO (kiến trúc + vận hành) · CFO (Phần 8 Pricing Concept, Phần 4 phí huỷ) · Legal (Phần 4 trách nhiệm bưu kiện/hàng cấm, Phần 11 bằng chứng giao hàng) — phê duyệt bắt buộc trước khi bất kỳ phần nào được triển khai
**Effective Date:** 2026-07-11
**Status:** THIẾT KẾ KIẾN TRÚC — không phải task code. Không sửa BRB, không sửa bất kỳ source code nào, không tạo proto, không build, không commit. Toàn bộ nội dung là đề xuất chờ duyệt.
**Nguồn sự thật khi có mâu thuẫn:** `docs/business/business-rule-bible-v1.0.md` (BRB) là SSOT vận hành cho Ride hôm nay. BRB **im lặng gần như hoàn toàn** về Delivery — Part 15 (§15.1 Food Delivery, §15.2 Parcel Delivery) chỉ là định hướng mở rộng, tự nhận *"these are not current business rules; they are design constraints that current rules must not contradict"* (BRB §15, dòng mở đầu Part 15). Vì vậy gần như mọi rule trong tài liệu này là **[MỚI]** (đề xuất chưa có trong BRB) hoặc **[MỚI — áp dụng loại suy từ BRB]** (mượn số/rule đã duyệt cho Ride, áp cho Delivery bằng suy luận tương tự), đúng theo yêu cầu của BRB Preamble: *"When this document is silent on a topic, a CPO-level decision is required before implementation proceeds."* Không có nội dung nào ở đây được coi là rule có hiệu lực cho tới khi qua quy trình tu chính chính thức (Constitution Article XI).
**Tài liệu đã đọc trước khi viết:** `docs/project/MVP_DEVELOPMENT_PLAN.md`; `docs/business/business-rule-bible-v1.0.md` (đọc toàn bộ 1974 dòng); `docs/business/mission/project-constitution-v0.1.md` (DOC-0001, đặc biệt Article VIII §8.1/§8.3); DOC-0002 (Product Vision) §6.16/§6.6/§6.21 (tham chiếu qua Constitution); `docs/driver/DRIVER_APP_SPEC.md`; toàn bộ `docs/business/*.md` còn lại (xác nhận không tài liệu nào đã thiết kế Delivery); mã nguồn `backend/services/trip/**`, `backend/services/dispatch/**`, `backend/services/booking/**`, `backend/services/pricing/**`, `backend/services/driver/**` (domain/app/infra/grpc, entity, state machine, proto); cấu trúc `apps/rider/**` và `apps/driver/**` (router, feature layering, trip lifecycle, shared widgets). Nghiên cứu ngoài (business flow/UX/pricing logic — không copy UI): GrabExpress, Ahamove, Lalamove, Be Delivery (beDelivery), Uber Connect, Xanh SM Express — nguồn liệt kê cuối tài liệu.

---

## TÓM TẮT ĐIỀU HÀNH

Delivery **không phải một dự án mới** — đây là một **TripType mới** (`ride` | `delivery`) cắm vào kiến trúc Trip/Dispatch/Booking/Pricing đã có, đúng theo mệnh lệnh kiến trúc đã được ghi thành văn từ trước: *"The architecture MUST support the multi-product future even though the MVP implements only the Ride product. Service boundaries, data models, and APIs MUST be designed to accommodate additional product lines without requiring rearchitecting of existing services."* (DOC-0001 Article VIII §8.1). Delivery được xếp **Phase 2** trong roadmap chính thức (DOC-0001 §8.1; DOC-0002 §6.16), tách biệt rõ khỏi Food Delivery và B2B Logistics (Phase 3).

Việc audit codebase (Trip/Dispatch/Booking/Pricing/Driver, cả 2 app Flutter) cho thấy kiến trúc hiện tại **đã sẵn sàng cho mở rộng theo hướng additive**: `Trip` không có khái niệm `TripType` nhưng cũng không hard-code giả định "chỉ chở người" ở tầng đặt tên (`PickupAddress`/`DropoffAddress`, không phải `PassengerSeat`); `Dispatch` đã trip-content-agnostic (chỉ quan tâm toạ độ điểm đón); `Booking` là một lớp orchestration thuần, không có entity/DB riêng, nên thêm luồng nghiệp vụ mới không đụng schema của chính nó; `Pricing` đã có Rule Engine config-driven (`PricingRule`/`RuleConfig`) với một mẫu "flat-fee rule" (Airport Fee) hoàn toàn có thể nhân bản cho phụ phí trọng lượng/quá khổ. Điểm cần bổ sung thật sự mới: cột phân loại `trip_type`, một entity mở rộng `DeliveryDetails` (1:1 với Trip), luồng OTP + bằng chứng giao hàng (hoàn toàn chưa tồn tại ở cả hai app), và một vài RPC/action mới ở tầng Booking/Driver app.

Tài liệu này thiết kế đúng **MVP** theo phạm vi được giao: một gói hàng/một chuyến, một chặng (không multi-stop), không COD, không Food Delivery, không xuyên biên giới, không drone/robot.

---

## PHẠM VI MVP (nhắc lại tường minh trước khi vào thiết kế)

**Nằm trong phạm vi:** một `TripType` mới (`delivery`) cắm vào Trip/Dispatch/Booking/Pricing hiện có; một gói hàng — một chuyến — một chặng (1 điểm lấy, 1 điểm giao); OTP + ảnh (tuỳ chọn) làm bằng chứng giao hàng; chính sách huỷ/thời gian chờ mượn loại suy từ BRB (Ride); danh mục hàng hoá config-driven; các thành phần giá (không kèm công thức).

**KHÔNG nằm trong phạm vi** (theo đúng yêu cầu, nhắc lại ngay từ đầu để không có thiết kế nào ở các phần sau vô tình lấn phạm vi):

| Loại trừ | Vì sao (đối chiếu roadmap) |
|---|---|
| Food Delivery (sản phẩm) | DOC-0001 §8.1 xếp Phase 3, tách biệt hoàn toàn khỏi Delivery (Phase 2) — dù "food" vẫn có thể là **nhãn phân loại gói hàng** trong Delivery (Phần 7), đây là hai khái niệm khác nhau |
| COD | Không có trong BRB, không có trong yêu cầu MVP — chỉ liệt kê như dòng khái niệm dự phòng (Phần 8, Phần 18) |
| Multi-stop / giao đa điểm | `dispatch.proto` hiện đã tự loại trừ "multi-order assignment"; giữ nguyên giới hạn 1 điểm lấy — 1 điểm giao |
| Xuyên biên giới / quốc tế | Ngoài phạm vi vận hành hiện tại (BRB §2.18 chỉ nói tới multi-city/multi-country cho Ride, không áp dụng ở đây) |
| Drone / Robot | Không có hạ tầng, không nằm trong bất kỳ tài liệu roadmap nào đã đọc |

---

## NGHIÊN CỨU THỊ TRƯỜNG (tham khảo business flow/pricing logic — không copy UI)

| Nền tảng | Business flow đáng chú ý | Pricing logic | Bằng chứng giao hàng | Chính sách đáng tham chiếu |
|---|---|---|---|---|
| **GrabExpress** (VN) | Điều hướng driver gần nhất trong ~15 phút, giao trong ~60 phút cho nội thành; 1 chuyến = 1 điểm lấy + 1 điểm giao | Giá cố định theo hướng (fixed fare by direction), hiển thị trước khi đặt | Ảnh chụp gửi cho người gửi khi hoàn tất — không thấy tài liệu công khai về OTP người nhận | SLA "trong vòng 24h" cho giao nội tỉnh nếu trễ |
| **Ahamove** | Nhiều gói theo tốc độ/loại xe: Super Fast (≤60 phút, tối đa 10 điểm giao), Super Fast Savings (1 điểm), Savings (2-4h); xe tải tới 2000kg | Công thức dạng: Base + Step×Distance + SoĐiểmDừng×PhíDừng — giá hiển thị trước khi xác nhận | Không tìm thấy chi tiết công khai | Liên tỉnh <30kg giao trong 3-4 ngày |
| **Lalamove** | Người gửi bật/tắt "Proof of Delivery" trong Settings; có tính năng "Delivery Note Return" gửi lại bằng chứng cho người gửi | Không tìm thấy công thức chi tiết công khai | **Ảnh HOẶC chữ ký** (driver chọn một trong hai), khuyến nghị ảnh có ngữ cảnh vị trí | Người nhận vắng mặt → gọi người gửi → thử trả hàng → hàng vô chủ giữ tại văn phòng 14 ngày rồi thanh lý; huỷ miễn phí trước khi ghép tài xế, ≤15 phút sau khi ghép nếu là đơn tức thời |
| **Be Delivery (beDelivery)** | Ghép tài xế gần như tức thời sau khi tạo đơn; có gói "2H" giá khởi điểm thấp hơn cho đơn ít gấp | Hỗ trợ COD (không nằm trong MVP này), phụ phí giờ khuya 23:00-06:00, giá đã gồm VAT, hiển thị giá tạm tính trước khi xác nhận | Không tìm thấy chi tiết công khai | Phụ phí động theo khu vực/khung giờ cao điểm |
| **Uber Connect** | 1 điểm lấy + 1 điểm giao, hầu hết đơn lấy-giao trong vòng 1 giờ | Giá trọn gói hiển thị trước (distance + demand + toll); giới hạn khối lượng/kích thước rõ ràng (5kg, 10×10×10 inch), danh sách hàng cấm rõ ràng (vũ khí, ma tuý, hàng giá trị cao) | Ảnh + tracking thời gian thực; phí chờ 0.35 USD/phút sau 2 phút miễn phí | Từ chối hàng giá trị cao (>200 USD tại Mỹ) — không có bảo hiểm mặc định |
| **Xanh SM Express** (VN) | Icon "Giao hàng" ngay trên app hiện có (giống mô hình beDelivery), hỗ trợ giao đa điểm + quay lại điểm nhận (return-to-sender) | Giá bậc thang đơn giản: cố định cho 2km đầu, phụ phí/km từ km tiếp theo; phụ phí "giao tận tay" (D2D) riêng | Không tìm thấy chi tiết công khai | Có gói COD (ngoài MVP này) |

**Rút ra cho thiết kế MVP (không copy UI, chỉ mượn logic nghiệp vụ):** (1) mọi nền tảng đều hiển thị giá **trước khi xác nhận** — MVP kế thừa nguyên tắc này (Phần 5, Phần 8); (2) bằng chứng giao hàng phổ biến nhất là **ảnh**, chữ ký là tuỳ chọn phụ — MVP chọn **OTP (bắt buộc) + Ảnh (tuỳ chọn)**, bỏ chữ ký (Phần 11); (3) chính sách "người nhận vắng mặt → gọi → thử trả lại" của Lalamove là khuôn mẫu tốt nhất cho trạng thái lỗi (Phần 3, Phần 6); (4) không nền tảng nào công khai chi tiết OTP người nhận — MVP tự thiết kế luồng này từ đầu (Phần 10), không có tiền lệ để mượn nguyên; (5) COD xuất hiện ở 3/6 nền tảng nhưng **bị loại khỏi phạm vi MVP theo yêu cầu** — chỉ liệt kê như dòng khái niệm ở Phần 8, không thiết kế.

Nguồn: [GrabExpress VN](https://www.grab.com/vn/en/express/) · [Grab SG Delivery Guide](https://www.grab.com/sg/ge-driver-deliveryguide/) · [Ahamove Estimate Order Fee API docs](https://developers.ahamove.com/en/docs/api-reference/order-apis/estimate-order-fee) · [Lalamove FAQ (HCMC)](https://www.lalamove.com/en-vn/faq) · [Lalamove blog — Proof of Delivery features](https://www.lalamove.com/en-ae/blog/interesting-lalamove-app-features-to-manage-your-deliveries-with-ease) · [beDelivery — Be](https://be.com.vn/en/consumer/be-delivery/) · [Uber Connect Guide 2026](https://therideshareguy.com/uber-connect/) · [Uber Connect FAQ](https://help.uber.com/driving-and-delivering/article/uber-connect---package-delivery-faq?nodeId=34d104e2-3ce3-444f-b5dc-cbcad889337d) · [Xanh SM Express](https://www.xanhsm.com/xanh-express)

---

## PHẦN 1 — KIẾN TRÚC TỔNG THỂ

Delivery **không** là một microservice riêng. Nó là một lát cắt xuyên suốt (cross-cutting slice) qua các service hiện có, theo nguyên tắc "thêm, không sửa" (additive-only) đã được kiểm chứng trong sprint Pricing V3 vừa xong của cùng codebase này:

```
Rider App                     Booking Service                  Trip Service
┌─────────────┐   BookDelivery   ┌──────────────────┐  CreateTrip     ┌───────────────┐
│ features/    │ ───────────────▶│ BookDeliveryUseCase│(trip_type=    │ trips (existing│
│ delivery/    │                  │  (mirror BookRide) │ delivery)  ──▶│ table +        │
└─────────────┘                  └─────────┬─────────┘                │ trip_type cột  │
                                            │ RequestDispatch          │ mới)           │
                                            ▼                          │ delivery_details│
                                  Dispatch Service (KHÔNG ĐỔI)          │ (bảng mới, 1:1)│
                                  offerNextDriver — vẫn chỉ quan        └───────┬────────┘
                                  tâm toạ độ điểm đón, không cần                │
                                  biết trip_type ở MVP                          │
                                            │                                   │
                                            ▼                                   ▼
Driver App                        Pricing Service                    ConfirmPickup /
┌─────────────┐  poll offer  ┌──────────────────────┐                ConfirmDelivery
│ TripPage mở  │◀─────────────│ WeightSurchargeRule /  │  (RPC mới, mirror
│ rộng, không  │              │ OversizeSurchargeRule  │   MarkDriverArrived/
│ tab riêng    │              │ (mirror AirportFeeRuleV3)│  StartTrip/CompleteTrip)
└─────────────┘              └──────────────────────┘
```

**Quyết định kiến trúc cốt lõi (4 điểm), mỗi điểm giữ nguyên hợp đồng của service liên quan:**

1. **`TripType`** — enum mới `ride | delivery`, sống trên `Trip` (cột mới `trip_type`, mặc định `'ride'` cho toàn bộ dữ liệu hiện có — 100% tương thích ngược). Đây là discriminator duy nhất cả hệ thống cần biết để phân biệt hai luồng.
2. **`DeliveryDetails`** — entity mở rộng 1:1 với Trip, **chỉ tồn tại khi `trip_type = delivery`**. Toàn bộ trường đặc thù giao hàng (người nhận, loại hàng, OTP, bằng chứng giao hàng) nằm ở đây — không thêm bất kỳ cột nào vào bảng `trips` hiện có ngoài `trip_type`. Cách làm này lặp lại chính mẫu hình đã dùng cho `AirportLeg`/`FullFareBreakdownV3` trong Pricing V3 (field mới, additive, không đụng struct cũ).
3. **Dispatch giữ nguyên 100% cho MVP** — vì Dispatch vốn dĩ đã trip-content-agnostic (chỉ cần `TripID`/`RiderID`/toạ độ điểm đón), không cần sửa gì để chạy được với Delivery. Giới hạn đã biết trước (không lọc theo loại xe/năng lực tài xế) là gap có sẵn từ trước, không phải gap Delivery gây ra — xem Phần 9.
4. **Pricing mở rộng bằng Rule Engine có sẵn**, không phải một pricing engine riêng — một `WeightSurchargeRule`/`OversizeSurchargeRule` mới implement đúng interface `PricingRule` đã có, giống hệt cách `AirportFeeRuleV3` đang hoạt động (`CategoryFlatFee`, tra cứu theo cấu hình YAML, không hard-code số).

**Về driver app "single-vertical":** `docs/driver/DRIVER_APP_SPEC.md` khẳng định rõ 3 lần (dòng 16/431/708) app tài xế hiện tại "🚫 not applicable — PandaDriver is single-vertical (ride only)", trong khi DOC-0002 §6.21 lại hình dung "tài xế chuyển đổi giữa Ride/Delivery". Thiết kế này **giải quyết mâu thuẫn** bằng cách chọn phương án không cần bật/tắt chế độ: tài xế nhận **cả hai loại cuốc qua cùng một hàng đợi offer** (unified inbox), phân biệt bằng icon/nhãn trên thẻ offer — không có công tắc "chế độ Delivery" nào trong MVP. Đây là quyết định cần CPO xác nhận rõ ràng là hướng mong muốn (Rủi ro, Phần 16).

Cách làm này chính là điều DOC-0001 Article VIII §8.1 yêu cầu: các TripType tương lai (Food/Intercity/Rental) chỉ cần thêm một giá trị enum + một bảng `*_details` mới + một bộ `PricingRule` mới — không đụng vào Trip/Dispatch/Booking hiện có.

---

## PHẦN 2 — DOMAIN MODEL

Đối chiếu với danh sách entity người dùng yêu cầu (Parcel, Package, Sender, Receiver, DeliveryTrip, DeliveryStatus, PackageCategory, OTP, ProofOfDelivery) — MVP **gộp một số khái niệm** để tránh entity thừa, mỗi lựa chọn được nêu lý do rõ ràng:

- **TripType** *(mới)* — enum mở, MVP chỉ có `ride`, `delivery`. Dự phòng tên cho tương lai: `food`, `intercity`, `rental` (không implement).
- **Trip** *(mở rộng additive)* — giữ nguyên toàn bộ field hiện có (`TripID, RiderID, DriverID, Status, PickupAddress, DropoffAddress, ...`), thêm duy nhất `TripType`. **`RiderID` được tái sử dụng làm "Sender"** cho Delivery (người đặt/người trả tiền) — không đổi tên cột (tránh breaking change), nhưng tầng API/UI phải alias rõ thành "Người gửi" để tránh nhầm lẫn dữ liệu/báo cáo (xem Rủi ro, Phần 16).
- **Sender** — **không phải entity riêng**. Là chính `Trip.RiderID` (người dùng đã đăng nhập, đặt đơn). Không cần bảng mới.
- **Receiver** *(mới, value object, KHÔNG phải tài khoản)* — tên, số điện thoại, ghi chú, địa chỉ (= `Trip.DropoffAddress`). MVP **không** xây dựng tài khoản/app cho người nhận (sẽ cần luồng đăng ký/auth riêng, ngoài phạm vi) — người nhận là dữ liệu do người gửi nhập, không xác thực danh tính ngoài OTP tại thời điểm giao (Phần 10).
- **Parcel / Package** — MVP coi đây là **một khái niệm duy nhất** ("Package"), không tách hai entity như tên gọi gợi ý. MVP là **một gói hàng / một chuyến** (không multi-parcel bundling), nên các thuộc tính gói hàng (category, size tier) nằm trực tiếp trên `DeliveryDetails`, không cần bảng `packages` chuẩn hoá riêng. Nếu tương lai hỗ trợ nhiều gói/chuyến, đây là điểm mở rộng rõ ràng (Phần 18).
- **DeliveryDetails** *(mới, 1:1 với Trip khi `trip_type=delivery`)* — chứa: `PackageCategory`, `PackageSizeTier` (tuỳ chọn cân nặng — xem Phần 7/8), `Note` (ghi chú cho tài xế), `ReceiverName`, `ReceiverPhone`, `ReceiverNote`, `Status` (xem `DeliveryStatus` bên dưới).
- **DeliveryTrip** — **không phải entity/bảng ghi riêng**. Đây là một khái niệm trình bày (read-model/DTO): sự kết hợp giữa `Trip{TripType=delivery}` và `DeliveryDetails`, trả về client dưới tên "DeliveryTrip" trong API response — tránh tạo nguồn sự thật thứ hai cho trạng thái chuyến đi.
- **DeliveryStatus** *(mới, enum)* — trạng thái chi tiết hơn `Trip.Status` (vốn khá thô: `searching/driver_assigned/driver_arrived/in_progress/completed/...`). Lớp phủ (overlay) lên trên, chi tiết đầy đủ ở Phần 3.
- **PackageCategory** *(mới, config-driven, KHÔNG hard-code)* — danh sách nhãn phân loại, nạp từ cấu hình (giống mẫu YAML của Pricing V3), chi tiết Phần 7.
- **OTP** *(mới, value object)* — `Code` (băm, không lưu plaintext), `GeneratedAt`, `ExpiresAt`, `VerifiedAt` (nullable), `AttemptCount`, `MaxAttempts`.
- **ProofOfDelivery** *(mới, value object)* — `Method` (`otp` bắt buộc, `photo` tuỳ chọn — không có `signature` ở MVP), `PhotoURL` (nullable), `GPSLat/GPSLon` (chụp tại thời điểm xác nhận), `Timestamp` (server-assigned), `ConfirmedByDriverID`.

**Bảng field chi tiết (mô tả khái niệm, không phải schema Go/SQL):**

| Entity | Field | Kiểu (khái niệm) | Bắt buộc? | Ghi chú |
|---|---|---|---|---|
| DeliveryDetails | PackageCategory | id tham chiếu config (Phần 7) | Có | Không phải free-text |
| DeliveryDetails | PackageSizeTier | id tham chiếu config (Phần 8) | Tuỳ chọn MVP | Ảnh hưởng phụ phí quá khổ |
| DeliveryDetails | Note | text | Không | Ghi chú cho tài xế |
| DeliveryDetails | ReceiverName | text | Có | |
| DeliveryDetails | ReceiverPhone | text (chuẩn hoá số VN) | Có | Đích gửi OTP (Phần 10) |
| DeliveryDetails | ReceiverNote | text | Không | Hướng dẫn giao hàng riêng |
| DeliveryDetails | Status | DeliveryStatus enum | Có | Xem Phần 3 |
| OTP | Code | chuỗi đã băm | Có | Không lưu plaintext |
| OTP | ExpiresAt | timestamp | Có | Gắn với vòng đời chuyến, không phải TTL cố định độc lập |
| OTP | AttemptCount / MaxAttempts | số nguyên | Có | Mặc định đề xuất MaxAttempts=3 (ASSUMPTION, Phần 10) |
| ProofOfDelivery | Method | `otp` \| `photo` | Có (ít nhất `otp`) | `photo` là bổ sung, không thay thế `otp` |
| ProofOfDelivery | GPSLat/GPSLon | số thực | Có | Ghi tại thời điểm `ConfirmDelivery`, không phải toạ độ điểm giao đã khai báo |

---

## PHẦN 3 — STATE MACHINE

`Trip.Status` (10 giá trị hiện có, không đổi) tiếp tục là trạng thái **thô** dùng chung cho cả Ride và Delivery. `DeliveryStatus` là lớp **chi tiết hơn**, chỉ có ý nghĩa khi `trip_type=delivery`, được `DeliveryDetails` theo dõi song song:

```
Searching ──▶ DriverAccepted ──▶ ArrivingPickup ──▶ PickedUp ──▶ Delivering ──▶ Delivered ──▶ Completed
    │               │                   │                                          │
    │               │                   │                                    RecipientUnavailable
    ▼               ▼                   ▼                                          │
Cancelled      Cancelled           Cancelled                              ┌────────┴────────┐
                                                                           ▼                  ▼
                                                                       Returning          (retry Delivering)
                                                                           │
                                                                           ▼
                                                                       Returned ──▶ Completed (billed)

(mọi giai đoạn tìm tài xế không thành công) ──▶ DriverTimeout (= Dispatch JobStatusFailed, xem Phần 9)
```

| DeliveryStatus | Trip.Status tương ứng | Kích hoạt bởi | Ghi chú / căn cứ |
|---|---|---|---|
| Searching | `searching` | Hệ thống, ngay sau `BookDelivery` | Giống hệt Ride |
| DriverAccepted | `driver_assigned` | Tài xế `AcceptDispatchOffer` | Giống hệt Ride |
| ArrivingPickup | `driver_assigned` (không đổi Trip.Status) | Chỉ là trạng thái hiển thị UI phía tài xế, không có transition backend riêng — giống cách `driverArriving` ở rider app hiện suy ra phía client, không lưu backend | [MỚI — áp dụng loại suy] |
| (Đã đến điểm lấy) | `driver_arrived` | `MarkDriverArrived` (RPC có sẵn, không đổi) | Tái sử dụng nguyên trạng |
| PickedUp | `in_progress` (qua `Trip.Start()`) | **`ConfirmPickup` — RPC mới** | Tương đương hành động "bắt đầu chuyến" của Ride, nhưng cần một xác nhận rõ ràng "đã cầm hàng" — Ride không có bước tương đương vì hành khách tự lên xe |
| Delivering | `in_progress` (không đổi) | Hệ thống, ngay sau PickedUp | Sub-status nội bộ trong `delivery_details`, không có transition Trip mới |
| Delivered | `completed` (qua `Trip.Complete()`) | **`ConfirmDelivery` — RPC mới**, yêu cầu OTP hợp lệ (Phần 10) | Tương đương `FinishTrip` của Ride nhưng có thêm điều kiện xác thực |
| RecipientUnavailable | `in_progress` (không đổi) | Tài xế báo không liên lạc được người nhận | [MỚI], mượn mẫu hình "gọi người gửi, thử trả hàng" từ Lalamove |
| Returning | `in_progress` (không đổi) | Tài xế xác nhận mang hàng quay lại sau khi RecipientUnavailable hết số lần thử | [MỚI] |
| Returned | `completed` (qua `Trip.Complete()` với cờ đặc biệt) | `ConfirmReturn` — RPC mới | Vẫn tính phí chuyến + phụ phí hoàn trả (Phần 8) — **chưa có rule billing hoàn trả trong BRB**, đây là [MỚI] cần CFO duyệt |
| Cancelled | `cancelled` | `CancelDelivery` (mirror `CancelRide`) | Tái sử dụng chính sách huỷ Ride theo loại suy (Phần 4) |
| DriverTimeout | Trip vẫn ở `searching` | Dispatch hết `MaxAttempts` (`JobStatusFailed`) | **Gap có sẵn từ Ride** — hiện Trip không có transition backend rõ ràng khi Dispatch fail hẳn; Delivery kế thừa nguyên gap này, không phải vấn đề mới |
| Completed | `settled` | Sau `PayDelivery`→`MarkPaid`→`Settle` (tái sử dụng `PayTripUseCase` nguyên trạng) | Giống hệt Ride |

---

## PHẦN 4 — BUSINESS RULES

Vì BRB im lặng gần như hoàn toàn, mỗi rule dưới đây được gắn nhãn rõ nguồn gốc.

- **Huỷ đơn phía người gửi** *[MỚI — áp dụng loại suy từ BRB §10.1]* — đề xuất tái sử dụng nguyên số đã duyệt cho Ride: miễn phí trong 2 phút đầu; sau đó, trước khi tài xế đến điểm lấy: phí 10.000đ (tài xế 80% = 8.000đ, platform 20% = 2.000đ); sau khi tài xế đã đến điểm lấy: phí 20.000đ (tài xế 16.000đ, platform 4.000đ). Đây là lựa chọn **ưu tiên nhất quán** hơn là tối ưu riêng cho Delivery — cần CFO xác nhận vì cấu trúc chi phí tài xế giao hàng (không có rủi ro "chở khách rồi bị huỷ giữa đường") khác với Ride.
- **Huỷ đơn phía tài xế miễn phạt** *[MỚI — áp dụng loại suy từ BRB §9.2]* — kế thừa 3 lý do đã duyệt cho Ride (điểm đến ngoài khu vực phục vụ, vị trí không thể tiếp cận, lo ngại an toàn) + **1 lý do mới cho Delivery**: hàng hoá không khớp khai báo hoặc thuộc danh mục cấm phát hiện tại điểm lấy — [MỚI], mượn tinh thần chính sách hàng cấm của Uber Connect.
- **Thời gian chờ tại điểm lấy hàng** *[MỚI — áp dụng loại suy từ BRB §2.2.9]* — tái sử dụng nguyên số Waiting Fee của Ride: 3 phút miễn phí, sau đó tính phí chờ; tối đa 10 phút chờ tổng (7 phút tính phí) thì tài xế được huỷ không bị phạt.
- **Thời gian chờ tại điểm giao hàng** *[MỚI — chưa có tiền lệ BRB]* — Ride không có khái niệm "chờ ở điểm đến" (hành khách xuống xe là xong), nên đây là rule hoàn toàn mới. Đề xuất tạm dùng cùng cấu trúc 3 phút miễn phí / phí sau đó / mốc tối đa trước khi được chuyển sang `RecipientUnavailable`, số cụ thể để CFO/COO quyết — không đề xuất số ở đây vì ngoài phạm vi loại suy an toàn.
- **Người nhận vắng mặt → trả hàng** *[MỚI]* — mượn mẫu Lalamove: tài xế gọi người gửi qua số đã đăng ký → nếu không xử lý được, chuyển `Returning` → giao lại tại điểm lấy ban đầu (không hỗ trợ điểm trả khác trong MVP, tránh phát sinh multi-stop). Số lần thử liên hệ tối đa và thời gian chờ cụ thể: ASSUMPTION, cần COO duyệt.
- **Danh mục hàng cấm** *[MỚI]* — chưa có trong BRB. Đề xuất danh sách khởi điểm dựa trên nghiên cứu Uber Connect/Lalamove: hàng hoá bất hợp pháp, vũ khí, chất cấm, vật liệu nguy hiểm, động vật sống, tiền mặt/trang sức giá trị cao không khai báo. **Bắt buộc Legal duyệt trước khi公bố công khai** — đây trực tiếp là gap mà chính BRB §15.2 đã tự nhận: *"Key business rule to define later: Liability for lost or damaged parcels. This requires an insurance product before launch."*
- **Trách nhiệm khi mất/hỏng hàng** *[MỚI — CHẶN LAUNCH CÔNG KHAI]* — BRB §15.2 đã tự flag chưa có sản phẩm bảo hiểm. MVP đề xuất: trách nhiệm platform = 0 (goodwill-only) cho tới khi có sản phẩm bảo hiểm — đây không chặn việc **thiết kế/build** (đúng phạm vi tài liệu này) nhưng chặn **launch công khai** (Rủi ro, Phần 16).
- **Ràng buộc loại hàng ↔ loại xe** *[MỚI]* — ví dụ hàng "Lớn" (Phần 7) chỉ nên được mời tới tài xế Van. Đây là rule mới cần thiết để Dispatch/matching hợp lý (dù MVP dispatch chưa tự động lọc — Phần 9).
- **COD** — **ngoài phạm vi MVP theo yêu cầu**, và BRB cũng không có rule nào để tham chiếu dù muốn — chỉ liệt kê như dòng khái niệm ở Phần 8/18, không thiết kế rule.

---

## PHẦN 5 — BOOKING FLOW

```
Người gửi (Trip.RiderID, đã đăng nhập)
   ↓
Điểm lấy hàng — tái sử dụng PlaceSearchField/PickupCard nguyên trạng
   ↓
Điểm giao hàng — tái sử dụng PlaceSearchField/DestinationCard nguyên trạng
   ↓
Thông tin người nhận (tên, SĐT — bắt buộc; đây là input MỚI không tồn tại ở Ride)
   ↓
Loại hàng — chọn từ PackageCategory config-driven (Phần 7)
   ↓
Ghi chú cho tài xế (tuỳ chọn — tái sử dụng pattern TextField optional có sẵn)
   ↓
[Ước tính giá hiển thị — itemised theo Phần 8, nhãn "ước tính" như Ride]
   ↓
Đặt → BookDeliveryUseCase (mirror BookRideUseCase):
        1. CreateTrip(trip_type=delivery)
        2. Tạo DeliveryDetails (package_category, receiver info, note)
        3. RequestDispatch(tripID, riderID, pickupLat, pickupLon) — KHÔNG ĐỔI so với Ride
        4. Saga compensation giống Ride: RequestDispatch lỗi → CancelTrip("dispatch_request_failed")
   ↓
Dispatch (Phần 9, không đổi) → Driver nhận offer → chấp nhận
   ↓
Pickup (ArrivingPickup → Đã đến điểm lấy → PickedUp qua ConfirmPickup)
   ↓
Delivery (Delivering → tài xế di chuyển tới điểm giao)
   ↓
OTP (Phần 10 — tài xế nhập mã người nhận đọc, xác thực qua ConfirmDelivery)
   ↓
Complete (Trip.Complete → InitiatePayment → PayDelivery → Settle, tái sử dụng nguyên luồng thanh toán Ride)
```

**Khoảng trống đáng chú ý kế thừa từ Ride:** audit xác nhận Booking hiện tại **không hề gọi** `PricingService.EstimateFare` trước khi đặt — rider đi thẳng từ nhập địa chỉ tới `BookRide`, chỉ có ước tính phía client (`MockFareBreakdown`). Với Delivery, vì loại hàng/kích cỡ ảnh hưởng giá rõ hơn nhiều so với việc chọn loại xe ở Ride, đề xuất **[MỚI, khuyến nghị không bắt buộc]**: nối `EstimateFare` thật vào luồng đặt Delivery ngay từ đầu, thay vì lặp lại gap ước tính-giả của Ride. Đây là khuyến nghị kiến trúc, không phải thiết kế thuật toán (đúng yêu cầu Phần 8 không viết công thức).

---

## PHẦN 6 — DRIVER FLOW

```
Online (không đổi — AvailabilityService hiện có)
   ↓
Nhận đơn — hàng đợi offer HỢP NHẤT (unified inbox), không có tab/chế độ Delivery riêng (Phần 1)
   ↓
Đi lấy — ArrivingPickup, tái sử dụng UI "đã đến điểm đón" hiện có, đổi nhãn "điểm lấy hàng"
   ↓
Xác nhận lấy — ConfirmPickup (RPC MỚI — không có tương đương ở Ride, vì hành khách tự lên xe
                còn tài xế phải chủ động xác nhận đã cầm gói hàng)
   ↓
Đi giao — Delivering, tái sử dụng route/map tracking hiện có
   ↓
OTP — nhập mã do người nhận đọc qua điện thoại (Phần 10), kèm chụp ảnh tuỳ chọn (Phần 11)
   ↓
Hoàn thành — ConfirmDelivery thành công → Trip.Complete() → luồng thanh toán/đánh giá tái sử dụng nguyên trạng Ride
```

So với luồng tài xế hiện có (`offerAvailable → acting → activeTrip → awaitingPayment → completed`), Delivery **tái sử dụng đúng khung state** nhưng chèn thêm **một bước xác nhận mới** (`ConfirmPickup`) không tồn tại ở Ride, và bước hoàn tất (`ConfirmDelivery`) có điều kiện xác thực OTP thay vì chỉ một hộp thoại xác nhận đơn giản như "Kết thúc chuyến đi?" của Ride.

---

## PHẦN 7 — PACKAGE CATEGORY

**Không hard-code** danh mục trong Go — nạp từ cấu hình (mirror mẫu `pricing_v3.default.yaml` đã dùng cho Pricing V3), cho phép vận hành thêm/sửa danh mục mà không cần deploy code. Mỗi mục cấu hình gồm: id, nhãn hiển thị, icon, ghi chú cảnh báo (nếu có, ví dụ "Medicine" gợi ý cần bảo quản cẩn thận), không mang giá trị định giá trực tiếp.

Danh mục khởi điểm đề xuất cho MVP (tham khảo, không phải danh sách đóng cứng — vận hành có thể sửa qua config):

| id | Nhãn | Ghi chú |
|---|---|---|
| `document` | Tài liệu | |
| `food` | Đồ ăn *(lưu ý: đây là NHÃN PHÂN LOẠI gói hàng gửi qua Delivery, khác hoàn toàn với sản phẩm "Food Delivery" — vốn bị loại khỏi phạm vi MVP theo Phần "Roadmap")* | |
| `medicine` | Thuốc | Cảnh báo bảo quản |
| `flower` | Hoa | Dễ vỡ/dễ hỏng |
| `gift` | Quà tặng | |
| `small_parcel` | Kiện nhỏ | |
| `medium_parcel` | Kiện vừa | |
| `large_parcel` | Kiện lớn | Gợi ý chỉ mời tài xế Van (Phần 4) |
| `other` | Khác | |

Category chỉ phục vụ **UX + sàng lọc hàng cấm** — kích thước/trọng lượng (yếu tố định giá thật sự) là một chiều dữ liệu **tách biệt** (`PackageSizeTier`, Phần 8), tránh nhầm lẫn "loại hàng là gì" với "hàng nặng/to bao nhiêu".

---

## PHẦN 8 — PRICING CONCEPT

*(Không thiết kế công thức — chỉ liệt kê thành phần, theo đúng yêu cầu phạm vi.)*

Kiến trúc tái sử dụng nguyên Rule Engine của Pricing V3 (`PricingRule`/`RuleConfig`, config-driven). Các thành phần cần có:

- **Pickup fee** — tương đương Base Fare hiện có, tái sử dụng khái niệm, không phải rule mới về mặt kiến trúc.
- **Distance** — tái sử dụng Distance Tier Engine đã build cho Pricing V3. MVP đề xuất đơn giản nhất: **dùng lại thẳng 3 khoá `VehicleType` hiện có** (car/motorcycle/van) làm khoá cấu hình cho Delivery, thay vì tạo keyspace hoàn toàn mới (`delivery_bike`, ...) — vì kích cỡ gói hàng vốn đã tương quan với năng lực chở của 3 loại xe này. Phương án tạo keyspace riêng để lại cho Phase 2 nếu cần độ chi tiết cao hơn.
- **Weight surcharge** *(rule mới)* — mirror chính xác mẫu `AirportFeeRuleV3` (flat-fee rule, `CategoryFlatFee`, tra cứu theo bậc cân nặng từ config).
- **Oversize surcharge** *(rule mới)* — cùng mẫu, kích hoạt theo `PackageSizeTier` thay vì cân nặng.
- **Waiting fee** — tái sử dụng khái niệm Waiting Fee hiện có, mở rộng áp dụng ở cả điểm lấy và điểm giao (Phần 4).
- **Priority delivery** *(rule mới, có thể hoãn Phase 2)* — phụ phí cho tốc độ giao nhanh hơn tiêu chuẩn, mượn ý tưởng từ các gói "Instant/Super Fast" đã khảo sát. MVP có thể chỉ ship **một mức tốc độ duy nhất** và hoãn phân tầng ưu tiên.
- **COD fee** *(chỉ liệt kê, ngoài phạm vi thiết kế)* — dòng khái niệm dự phòng cho Phase 2, không có rule/con số nào ở đây.

---

## PHẦN 9 — DISPATCH

**MVP: không đổi gì so với Ride.** `RequestDispatch`/`offerNextDriver` (bán kính, timeout, số lần thử lại) chạy nguyên trạng, vì Dispatch vốn đã không đọc nội dung chuyến đi — chỉ cần `TripID`/`RiderID`/toạ độ điểm đón.

Hai khác biệt **được xác định nhưng hoãn sang Phase 2** (không chặn MVP):

1. **Lọc theo năng lực xe** — hiện `offerNextDriver` không lọc theo loại xe/khả năng chở cho **bất kỳ** loại chuyến nào (kể cả Ride) — đây là gap có sẵn từ trước, Delivery chỉ làm nó rõ ràng hơn (ví dụ gói hàng "Lớn" bị mời tới tài xế xe máy là vô lý). Đề xuất Phase 2 bổ sung điều kiện lọc trong `offerNextDriver` hoặc một bước tra cứu năng lực xe từ Driver service.
2. **Loại trừ tài xế đang bận** — `offerNextDriver` hiện chỉ loại trừ theo `HasBeenOffered`/trạng thái online (`IsActive`), **không** kiểm tra "tài xế đang có chuyến khác". Đây cũng là gap có sẵn ảnh hưởng cả Ride-chồng-Ride, nhưng hậu quả với Delivery cụ thể hơn (tài xế đang cầm gói hàng dở dang mà bị mời sang cuốc Ride khác). Khuyến nghị xử lý ở tầng nền tảng (không riêng cho Delivery), Phase 2.

**Không có dispatch đa đơn/nhiều điểm dừng ở MVP** — khớp với việc chính `dispatch.proto` đã tự loại trừ rõ "multi-order assignment" và khớp với giới hạn "không multi-stop" mà phạm vi MVP yêu cầu.

---

## PHẦN 10 — OTP FLOW

- **Khi nào sinh OTP:** đề xuất sinh **tại thời điểm `ConfirmPickup`** (không phải lúc đặt đơn) — lý do: tránh mã bị "nguội" nếu chuyến kéo dài, đúng tinh thần just-in-time. *[MỚI — không có tiền lệ công khai từ các nền tảng đã khảo sát để mượn nguyên]*.
- **Gửi mã cho ai:** vì Receiver **không có tài khoản** trong MVP (Phần 2), mã phải gửi qua **SMS trực tiếp tới `ReceiverPhone`** đã khai báo lúc đặt đơn — đây là một **phụ thuộc hạ tầng mới** (nhà cung cấp SMS gateway), chưa từng xuất hiện trong bất kỳ service nào đã khảo sát trong session này (Rủi ro, Phần 16).
- **Khi nào verify:** tài xế nhập mã (do người nhận đọc trực tiếp) vào app tài xế, gửi qua `ConfirmDelivery` (RPC mới) — server xác thực với mã đã băm lưu ở `DeliveryDetails`/bảng OTP riêng (Phần 14).
- **Nếu sai OTP:** cho phép tối đa **3 lần thử** *[MỚI/ASSUMPTION — không có số BRB để tham chiếu]*; hết số lần thử, app tài xế đưa ra lựa chọn dự phòng: gọi người gửi, hoặc chuyển `RecipientUnavailable`. **Không có cơ chế "bỏ qua OTP, hoàn tất thủ công" trong MVP** — giữ tính toàn vẹn xác thực, khác với ảnh (Phần 11) vốn chỉ là bằng chứng phụ.
- **Nếu mất mạng:** xác thực OTP bắt buộc cần kết nối server (kiểm tra phía server, không thể xác thực offline vì mã băm nằm trên server). Đề xuất app tài xế xếp hàng đợi lệnh `ConfirmDelivery(otp)` giống các hành động chuyến đi khác và tự động gửi lại khi có mạng trở lại — nhưng **chuyến giao hàng không được coi là hoàn tất phía client cho tới khi có xác nhận từ server**, tránh trạng thái "đã giao" giả trong lúc mất kết nối.

---

## PHẦN 11 — PROOF OF DELIVERY

Thiết kế: **OTP (bắt buộc) + Ảnh (tuỳ chọn)**. **Không có chữ ký (signature) trong MVP** — cả hai app hiện tại đều chưa có widget chữ ký hay bất kỳ hạ tầng camera/OTP nào (audit xác nhận không có dependency `image_picker`/`signature` ở cả hai app), nên đây là lựa chọn thu hẹp phạm vi có chủ đích, không phải thiếu sót — mượn đúng phần "ảnh" trong lựa chọn "ảnh HOẶC chữ ký" mà Lalamove cung cấp, bỏ nhánh chữ ký để giảm phụ thuộc mới.

- **Ảnh** *(tuỳ chọn)* — chụp tại thời điểm `ConfirmDelivery`, lưu trữ cần **object storage mới** (S3-compatible hoặc tương đương) — hạ tầng hiện tại (Postgres/Redis) không có chỗ chứa file nhị phân (Rủi ro, Phần 16).
- **Chữ ký** — **loại khỏi MVP**, ghi nhận là điểm mở rộng Phase 2.
- **OTP** — xem Phần 10, là yếu tố xác thực chính.
- **GPS** — toạ độ tài xế tại thời điểm gửi `ConfirmDelivery`, tái sử dụng hạ tầng vị trí đã có (`core/location`/tương đương phía driver app), không cần cơ chế thu thập vị trí mới.
- **Timestamp** — do server gán (không dùng đồng hồ client), nhất quán với cách `Trip.UpdatedAt` hiện đang hoạt động.

---

## PHẦN 12 — RIDER UI FLOW

Toàn bộ màn hình mới đề xuất theo đúng layering đã quan sát (`data/` + `domain/models/` + `presentation/{pages,widgets}`), đặt tại `features/delivery/` (mirror `features/booking/` + `features/trip/`):

1. **Điểm vào** — đề xuất **tab thứ 5 trên bottom nav** ("Giao hàng", ví dụ icon `Icons.local_shipping_outlined`) — khớp đúng mẫu hình mỗi feature-một-tab đã có (Home/Đặt xe/Ví/Hồ sơ), thay vì một card mới trên Home (Home hiện thuần bản đồ, không có tiền lệ card thứ cấp).
2. **`DeliveryPage`** (mirror `BookingPage`) — form đặt hàng full-page, cũng có thể mở dạng bottom-sheet từ bản đồ (mirror `booking_bottom_sheet.dart`).
3. **Chọn điểm lấy/điểm giao** — tái sử dụng nguyên `PlaceSearchField`/`PickupCard`/`DestinationCard`/`RouteConnector`, chỉ đổi nhãn.
4. **Form thông tin người nhận** *(mới)* — tên + SĐT, các trường text đơn giản mirror style `TextField` đã dùng cho ô nhận xét.
5. **Chọn loại hàng** — tái sử dụng đúng mẫu `VehicleSelector` (dãy thẻ cuộn ngang), đổi nguồn dữ liệu từ `VehicleOption` sang `PackageCategory` (Phần 7).
6. **Ghi chú** — tái sử dụng pattern ô nhận xét tuỳ chọn có sẵn.
7. **Thẻ ước tính giá** — tái sử dụng `FareSummaryCard`/`PriceBreakdownSheet`, liệt kê theo các thành phần Phần 8, vẫn gắn nhãn "ước tính".
8. **Nút xác nhận đặt** — tái sử dụng `AppButton`/mẫu `BookRideButton`.
9. **`DeliveryLifecyclePage`** *(mới, mirror `TripLifecyclePage`)* — cùng cơ chế poll 5s + `AnimatedSwitcher` theo `DeliveryStatus`: Đang tìm tài xế (tái sử dụng `searching_driver_view` pattern) → Đã nhận đơn/Đang đến lấy hàng (tái sử dụng `driver_assigned_view`/`driver_arriving_view`, đổi nhãn "tài xế" → có thể giữ nguyên vì vẫn là cùng một tài xế) → Đã lấy hàng/Đang giao (tái sử dụng `trip_in_progress_view`, thêm hiển thị tracking bản đồ) → Đã giao (màn hình mới: mascot ăn mừng + tóm tắt bằng chứng giao hàng — ảnh thu nhỏ nếu có + thời gian) → Thanh toán (tái sử dụng nguyên `_PaymentPendingView`) → Đánh giá (tái sử dụng nguyên mẫu 5 sao + nhận xét, đánh giá tài xế giao hàng).
10. **Lịch sử** — mở rộng `trip_history_page.dart` với bộ lọc/nhãn loại chuyến (Ride/Delivery), `trip_detail_page.dart` thêm phần hiển thị bằng chứng giao hàng khi `trip_type=delivery`.

---

## PHẦN 13 — DRIVER UI FLOW

1. **Hàng đợi offer hợp nhất** — tái sử dụng nguyên `_PollingView`, không đổi.
2. **Thẻ offer** — mirror `_OfferCard`: hiển thị loại hàng (icon/nhãn category) thay vì thông tin hành khách, cùng mẫu `_AddressRow`, `_InfoChip` khoảng cách/giá, đếm ngược, nút Nhận/Từ chối.
3. **Màn hình thực thi chuyến giao hàng** — mirror `_TripExecutionCard`: thêm **`DeliveryTimeline`** *(mới, mirror `TripTimeline`)* với 4 mốc thay vì 3 (pickup → pickedUp → delivering → delivered); thẻ thông tin người gửi (mirror `PassengerInfoCard`); **thẻ thông tin người nhận** *(mới)* — tên/SĐT/địa chỉ/ghi chú; nút hành động đơn, đổi nhãn tuần tự: "Tôi đã đến điểm lấy hàng" → "Xác nhận đã lấy hàng" (gọi `ConfirmPickup`) → "Tôi đã đến điểm giao hàng" → "Giao hàng" (mở màn hình xác nhận bằng chứng giao hàng).
4. **Màn hình xác nhận giao hàng** *(mới)* — ô nhập OTP (`AppOtpField`, widget dùng chung mới), nút chụp ảnh tuỳ chọn (cần thêm dependency `image_picker`), nút xác nhận cuối cùng qua `AppDialog.confirm` (mirror "Kết thúc chuyến đi?") → gọi `ConfirmDelivery`.
5. **Luồng người nhận vắng mặt** *(mới)* — hành động phụ xuất hiện trong lúc Delivering ("Không liên lạc được người nhận"), mở `AppDialog` dẫn tới: gọi người gửi (deep-link gọi điện) → chuyển `Returning`.
6. **Chờ thanh toán / hoàn tất** — tái sử dụng nguyên `_AwaitingPaymentCard`/`_TripCompletedCard`, đánh giá người gửi.

---

## PHẦN 14 — DATABASE CONCEPT

*(Chỉ mô tả entity/quan hệ — không viết SQL.)*

- **`trips`** *(bảng hiện có)* — thêm đúng **một cột mới**: `trip_type`, kiểu chuỗi, mặc định `'ride'`. Additive, không đổi/xoá cột nào hiện có — toàn bộ dữ liệu Ride cũ không bị ảnh hưởng.
- **`delivery_details`** *(bảng mới)* — quan hệ 1:1 với `trips` (khoá ngoại `trip_id`, duy nhất), chứa: loại hàng, kích cỡ/khối lượng, ghi chú, tên/SĐT/ghi chú người nhận, trạng thái (`DeliveryStatus`), thời gian tạo/cập nhật.
- **`delivery_otp`** *(bảng mới, tách riêng để có lịch sử thử mã — phục vụ audit)* — khoá ngoại `trip_id`, mã đã băm (không lưu plaintext — lưu ý bảo mật **[MỚI]**), thời điểm sinh/hết hạn/xác thực, số lần thử.
- **`delivery_proof`** *(bảng mới)* — khoá ngoại `trip_id`, phương thức, URL ảnh (nullable), toạ độ GPS, tài xế xác nhận, thời điểm.
- **Quan hệ:** `trips (1) — (1) delivery_details`, `trips (1) — (1) delivery_otp`, `trips (1) — (1) delivery_proof`.
- **Không cần** thêm cột/bảng nào ở `driver_profiles`, `vehicles`, hay `dispatch_jobs` cho MVP (lọc năng lực xe hoãn Phase 2, Phần 9) — giữ bề mặt migration tối thiểu.

---

## PHẦN 15 — API CONCEPT

*(Chỉ mô tả service nào chịu trách nhiệm gì — không viết proto.)*

| Service | Thay đổi cần thiết cho MVP |
|---|---|
| **Booking** | RPC mới: `BookDelivery` (mirror `BookRide`), `ConfirmPickup`, `ConfirmDelivery` (nhận OTP + ảnh tuỳ chọn), `MarkRecipientUnavailable`, `ConfirmReturn`. Tái sử dụng nguyên: huỷ (mirror `CancelRide`), thanh toán (mirror `PayRide`), chi tiết/danh sách đơn (mirror `GetBookingDetails`/`ListRiderTrips`/`ListDriverTrips`). |
| **Trip** | Thêm field `trip_type` (additive) trên request/response tạo chuyến — không phải RPC mới. Là nơi hợp lý nhất để sở hữu 3 bảng mới (`delivery_details`/`delivery_otp`/`delivery_proof`) vì đã sở hữu aggregate Trip + database liên quan — tránh phải dựng service/DB mới chỉ vì Delivery. |
| **Dispatch** | **Không cần đổi API** cho MVP (Phần 9). |
| **Pricing** | Thêm field tuỳ chọn (kích cỡ/khối lượng, khoá loại hàng) trên `EstimateFareRequest`/`CalculateFinalFareRequest` — additive, không breaking. Không cần RPC mới. |
| **Driver** | **Không cần đổi** cho MVP. Field `AcceptedTripTypes` (tuỳ chọn, Phase 2) không chặn MVP (Phần 1 — unified inbox mặc định nhận cả hai loại). |
| **Gateway (HTTP)** | Route mới `/api/v1/deliveries/**`, mirror cấu trúc `/api/v1/rides/**` hiện có. |

---

## PHẦN 16 — RISK

| Rủi ro | Ảnh hưởng | Ghi chú/Giảm thiểu |
|---|---|---|
| BRB im lặng hoàn toàn về Delivery | Toàn bộ rule trong tài liệu là ASSUMPTION — triển khai trước khi CPO/CFO/Legal duyệt là rủi ro tuân thủ nội bộ | Không code/launch trước khi có phê duyệt chính thức (đúng Status ở đầu tài liệu) |
| Trách nhiệm bưu kiện mất/hỏng chưa có bảo hiểm | BRB §15.2 tự nhận thiếu sản phẩm bảo hiểm — launch công khai trước khi có là rủi ro tài chính/pháp lý trực tiếp | Chặn **launch công khai**, không chặn build/thử nghiệm nội bộ |
| Người nhận không có tài khoản → OTP phải gửi SMS ngoài hệ thống | Phụ thuộc nhà cung cấp SMS gateway chưa khảo sát | Cần lựa chọn nhà cung cấp trước khi implement Phần 10 |
| Ảnh bằng chứng giao hàng cần object storage | Hạ tầng hiện tại (Postgres/Redis) không có | Cần chọn giải pháp lưu trữ (S3-compatible) trước khi build Phần 11 |
| Hai enum `VehicleType` độc lập đã tồn tại từ trước (Pricing vs Driver service) | Thêm khoá Delivery có nguy cơ lệch enum sang phiên bản thứ ba nếu không cẩn thận | Khuyến nghị dùng lại đúng 3 giá trị hiện có thay vì tạo keyspace mới (Phần 8) |
| Dispatch hiện không loại trừ tài xế đang bận khi mời offer mới | Gap có sẵn từ Ride, hậu quả rõ hơn với Delivery (tài xế đang cầm hàng dở lại nhận cuốc Ride) | Ghi nhận là gap nền tảng, khuyến nghị xử lý ở tầng Dispatch chung, Phase 2 |
| Xung đột tầm nhìn: DOC-0002 §6.21 hình dung tài xế "chuyển đổi chế độ", nhưng `DRIVER_APP_SPEC.md` khẳng định "single-vertical" | Quyết định "hàng đợi hợp nhất, không toggle" ở Phần 1 giải quyết xung đột nhưng là một lựa chọn, không phải sự thật hiển nhiên | Cần CPO xác nhận rõ ràng hướng này trước khi build |
| `RiderID` mang nghĩa kép "hành khách" (Ride) / "người gửi" (Delivery) | Rủi ro nhầm lẫn dữ liệu/báo cáo phân tích nếu không alias rõ ở tầng API/UI | Không đổi tên cột DB (breaking), chỉ alias tầng API/UI |
| Danh mục hàng cấm chưa chính thức, chưa có cơ chế kiểm tra ngoài "tài xế tự đánh giá" | Rủi ro vận hành/pháp lý khi phát sinh tranh chấp | Cần Legal duyệt danh sách trước khi launch công khai |

---

## PHẦN 17 — MIGRATION

- Migration mới tiếp theo dãy số đã có (`001_identity` … `007_idempotency`) → đề xuất **`008_delivery.up/down.sql`**, đúng quy ước đặt tên đang dùng.
- Toàn bộ thay đổi là **additive**: thêm cột có giá trị mặc định (`trips.trip_type`), thêm bảng mới (`delivery_details`, `delivery_otp`, `delivery_proof`) — không đổi kiểu, không xoá cột nào trên `trips`/`dispatch_jobs`/`driver_profiles`. Ride không bị ảnh hưởng bởi migration này.
- Không cần backfill dữ liệu cũ — toàn bộ `trips` hiện có tự động nhận `trip_type='ride'` qua giá trị mặc định của cột.
- Đề xuất rollout qua feature flag, theo đúng tiền lệ `PRICING_VERSION` đã dùng cho Pricing V3 — ví dụ `DELIVERY_ENABLED` ở Booking/Gateway, mặc định `false`, bật dần theo thành phố/nhóm tài xế thí điểm.

---

## PHẦN 18 — ROADMAP

**MVP (tài liệu này):** Ride + Delivery cùng kiến trúc, một gói hàng/một chuyến, một chặng, dispatch không lọc năng lực xe, bằng chứng giao hàng = OTP (bắt buộc) + ảnh (tuỳ chọn), chính sách huỷ mượn nguyên từ Ride, không COD, không bảo hiểm (chặn launch công khai, không chặn build).

**Phase 2 (theo DOC-0001 Article VIII §8.1, cùng nhịp Delivery):** COD; giao đa điểm/multi-stop; gói ưu tiên tốc độ (Priority delivery); bằng chứng giao hàng bổ sung chữ ký; `AcceptedTripTypes` cho tài xế + lọc năng lực xe ở Dispatch; sản phẩm bảo hiểm bưu kiện (mở khoá launch công khai); tài khoản/thông báo cho người nhận.

**Phase 3 (theo roadmap chính thức):** Food Delivery, B2B Logistics, tích hợp Merchant Platform.

**KHÔNG LÀM Ở MVP (nhắc lại theo đúng phạm vi được giao):** Food Delivery (sản phẩm), COD, giao đa điểm, xuyên biên giới/quốc tế, drone, robot.

---

## PHỤ LỤC — ĐỐI CHIẾU NHANH RIDE vs DELIVERY

| Khía cạnh | Ride (hiện có) | Delivery (thiết kế MVP) | Có đổi Ride không? |
|---|---|---|---|
| TripType | ngầm định, không có field | field `trip_type` mới, mặc định `'ride'` | Không (additive) |
| Actor chính | Rider + Driver | Sender (=RiderID) + Driver + Receiver (không tài khoản) | Không |
| Entity mở rộng | Không có | `DeliveryDetails`/`delivery_otp`/`delivery_proof`, 1:1 với Trip | Không |
| State machine thô | `Trip.Status` 10 giá trị | Dùng chung nguyên trạng | Không |
| State machine chi tiết | Không có lớp phủ | `DeliveryStatus` mới, lớp phủ trên `Trip.Status` | Không |
| Xác nhận bắt đầu | `Start()` (hành khách tự lên xe) | `ConfirmPickup` (RPC mới, tài xế chủ động xác nhận) | Không, chỉ thêm RPC |
| Xác nhận hoàn tất | `FinishTrip` (không cần xác thực) | `ConfirmDelivery` (bắt buộc OTP hợp lệ) | Không, chỉ thêm RPC |
| Dispatch matching | Theo bán kính, không lọc năng lực xe | Giống hệt Ride ở MVP (gap lọc năng lực xe hoãn Phase 2) | Không |
| Pricing | Distance Tier + Commission (Pricing V3) | Thêm Weight/Oversize surcharge rule, dùng lại keyspace VehicleType | Không (rule mới, không sửa rule cũ) |
| Huỷ đơn | BRB §10.1 | Mượn nguyên số BRB §10.1 theo loại suy | Không |
| Bằng chứng hoàn tất | Không có (chỉ trạng thái `completed`) | OTP bắt buộc + ảnh tuỳ chọn | Không |
| App tài xế | Đơn-vertical, 1 hàng đợi offer | Vẫn 1 hàng đợi offer, hiển thị cả 2 loại (không toggle mới) | Không |
| Migration | — | `008_delivery.up/down.sql`, additive only | Không có cột/bảng Ride nào bị đổi |

---

*Xác nhận cuối tài liệu: không có source code nào bị sửa trong quá trình thực hiện task này; không có proto nào được tạo; không có lệnh build/test/commit nào được chạy.*
