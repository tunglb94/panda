# Panda — Economy Engine (Thiết kế Kiến trúc Nghiệp vụ)

**Document Classification:** Internal — Confidential
**Authority:** CPO · CFO · CTO
**Effective Date:** 2026-07-10
**Status:** Draft v0.1 — tài liệu thiết kế kiến trúc nghiệp vụ, **chưa có dòng code nào hiện thực hoá** (xác nhận qua `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.1: Wallet/Payment/Promotion hiện là skeleton rỗng — 0 domain logic)
**Nguồn sự thật khi có mâu thuẫn:** `docs/business/business-rule-bible-v1.0.md` (BRB). Tài liệu này **không được mâu thuẫn** với BRB — mọi quy tắc đã có trong BRB được tham chiếu nguyên trạng; mọi khái niệm mới được đánh dấu **[MỚI]** và phải đi qua quy trình tu chính BRB trước khi triển khai.
**Tài liệu đã đọc trước khi viết:**
- `docs/business/PRICING_STRATEGY.md` (viết tắt **PS**) — chiến lược giá theo giai đoạn tăng trưởng
- `docs/business/PRICING_SIMULATION_REPORT.md` (viết tắt **PSR**) — kết quả mô phỏng 111 kịch bản, phát hiện các ASSUMPTION (VAT, Insurance) và gap cấu trúc (xe máy chuyến ngắn lỗ, Long Pickup Compensation tốn kém)
- `docs/business/business-rule-bible-v1.0.md` (viết tắt **BRB**) — nguồn sự thật vận hành: Part 2 (Pricing), Part 3 (Promotion), Part 4 (Voucher), Part 5 (Wallet), Part 6 (Settlement), Part 7 (Driver Economy), Part 8 (Driver Incentive), Part 9 (Driver Performance), Part 10 (Rider Rules), Part 11 (Fraud), Part 12 (Risk Engine), Part 14 (Financial Reports)
- `docs/project/MVP_DEVELOPMENT_PLAN.md` (viết tắt **MVP**) — xác nhận: Wallet/Payment/Promotion là skeleton rỗng; ADR-0014 (phương thức thanh toán) và UBD-001 (chấp nhận tiền mặt) đều **chưa có quyết định**

**Phạm vi:** Tài liệu business-level thuần tuý. Không code, không pseudo-code, không SQL, không Go, không Dart, không API, không protobuf, không DB schema. Không sửa file nào khác ngoài chính tài liệu này.

---

## TÓM TẮT ĐIỀU HÀNH

Economy Engine là bản thiết kế **toàn bộ nền kinh tế của Panda** — không chỉ "tính giá bao nhiêu" (đã có ở PS/PSR) mà là **mỗi đồng tiền đi đâu, từ đâu đến, chia thế nào, ai giữ, giải ngân khi nào**. Bốn nguyên tắc tài chính xuyên suốt (PHẦN 1): tăng trưởng, thanh khoản, công bằng, minh bạch — **không phải lợi nhuận tối đa**.

Tài liệu thiết kế 6 loại "ví"/quỹ (PHẦN 3), một Settlement Engine dạng state machine có rollback/retry (PHẦN 4), ba engine cấu hình-được (Commission/Promotion/Driver Incentive — PHẦN 5–7), hệ thống Membership hai phía (PHẦN 8), 8 nhóm chống gian lận (PHẦN 9), bộ KPI tài chính đầy đủ (PHẦN 10), nguyên tắc Rule Engine "không hardcode" áp dụng cho mọi engine (PHẦN 11), roadmap bật module theo quy mô rider (PHẦN 12), và danh sách những điều không nên làm (PHẦN 13).

Cuối tài liệu là danh sách **9 đề xuất cần bổ sung vào BRB** để hai tài liệu hợp nhất thành một nguồn duy nhất trong tương lai.

---

# PHẦN 1 — TRIẾT LÝ TÀI CHÍNH

## 1.1 Panda không tối đa hoá lợi nhuận

Đây không phải là một tuyên bố marketing — nó là một ràng buộc kiến trúc. Mọi engine trong tài liệu này (Commission, Promotion, Incentive, Membership) đều có một "van an toàn" ngăn nó tự tối ưu về phía lợi nhuận nếu điều đó đi ngược lại thứ tự ưu tiên đã thiết lập tại PS §1: (1) tăng Rider, (2) tăng Driver, (3) tăng tần suất, (4) tăng cạnh tranh, (5) hoà vốn. Lợi nhuận là **hệ quả**, không phải **mục tiêu tối ưu hoá** của bất kỳ Rule Engine nào trong hệ thống này.

## 1.2 Bốn trục tối đa hoá

### Tăng trưởng
Đo bằng số Rider hoạt động, số Driver hoạt động, số thành phố — không đo bằng doanh thu. Mọi Promotion/Incentive Engine (PHẦN 6–7) đều có chỉ số thành công gắn với tăng trưởng (CPIR — Cost Per Incremental Ride, đã định nghĩa tại BRB §3.8), không gắn với doanh thu ngắn hạn.

### Thanh khoản (Liquidity)
Panda là một **marketplace hai chiều**: giá trị của nền tảng với Rider phụ thuộc vào việc có đủ Driver gần đó (ETA thấp), và giá trị với Driver phụ thuộc vào việc có đủ Rider đặt xe (thu nhập ổn định). Thanh khoản là chỉ số "một Rider mất bao lâu để tìm được xe" và "một Driver mất bao lâu để có chuyến tiếp theo". Dynamic Surge (BRB §2.13, PS §5) là công cụ cân bằng thanh khoản theo thời gian thực — **không phải công cụ tối đa hoá doanh thu**, dù về mặt kế toán nó tạo thêm doanh thu. Toàn bộ kiến trúc ví/settlement trong tài liệu này được thiết kế để **không bao giờ làm chậm thanh khoản** — ví dụ: Driver Wallet không giữ tiền lâu hơn mức cần thiết để xác minh giao dịch hợp lệ (PHẦN 3.7), vì tiền bị giữ lâu = driver mất động lực tiếp tục hoạt động = thanh khoản giảm.

### Công bằng
Ba chiều công bằng đã xác lập tại BRB §1.3 (Rider Fairness / Driver Fairness / Platform Fairness) là nền tảng cho mọi thiết kế ví/settlement trong tài liệu này. Hệ quả kiến trúc trực tiếp: **Passenger Membership (PHẦN 8) không bao giờ thay đổi công thức giá theo hạng thành viên** — hai rider cùng chuyến, cùng điều kiện, phải trả cùng giá (Constitution Article II §2.1) bất kể hạng thành viên. Membership chỉ tạo khác biệt ở **dịch vụ** (ưu tiên ghép chuyến, hỗ trợ nhanh hơn), không ở **giá**.

### Minh bạch
Kế thừa trực tiếp BRB §1.2 Nguyên tắc 1–2 (Transparency Before Conversion, Rules Are Public). Hệ quả kiến trúc: mọi Wallet (PHẦN 3) phải có **ledger bất biến, append-only** (đã là quy tắc tại BRB §5.2/§5.12 cho Rider Wallet — tài liệu này áp dụng nguyên tắc y hệt cho Driver Wallet, Platform Wallet, và mọi quỹ khác, không có ngoại lệ).

---

# PHẦN 2 — MONEY FLOW

## 2.1 Sơ đồ tổng quát

```
Khách (Rider)
   │  [thanh toán: thẻ / ví Panda / tiền mặt — phương thức chưa chốt, MVP ADR-0014 unresolved]
   ▼
Payment Gateway (đối tác bên thứ ba, PCI-DSS)
   │  [Panda KHÔNG BAO GIỜ lưu số thẻ/CVV — BRB §9.3]
   ▼
Panda (tài khoản merchant/ngân hàng của nền tảng)
   │  [tiền vào tài khoản Panda ở dạng GROSS — chưa chia]
   ▼
Ví tạm giữ (Trip Escrow — trạng thái "Locked Balance", BRB §5.10)
   │  [tiền bị khoá từ lúc XÁC NHẬN ĐẶT XE, không phải lúc hoàn thành — vì chuyến có thể bị huỷ]
   ▼
── Settlement Engine kích hoạt khi Trip Finished (PHẦN 4) ──
   │
   ├──▶ Commission (phần giữ lại của Panda — vào Platform Wallet)
   ├──▶ Voucher (giảm trừ đã áp — không có dòng tiền vật lý mới, chỉ là doanh thu bị giảm, ghi nhận vào Promotion Fund)
   ├──▶ Khuyến mãi (Cashback/Referral Reward — CÓ dòng tiền thật, từ Promotion Fund → Passenger Wallet, dạng credit)
   ├──▶ Thuế (VAT — tách riêng vào Tax Holding, KHÔNG BAO GIỜ trộn với doanh thu Panda)
   ▼
Driver Wallet (trạng thái "Pending" → "Available" theo chu kỳ settlement, BRB §6.9)
   │  [chu kỳ chuẩn: hàng tuần; driver hạng Gold+ có thể rút nhanh hơn, có phí]
   ▼
Rút tiền (Withdrawal — driver chủ động yêu cầu)
   │  [T+1 ngày làm việc, xử lý bởi Panda]
   ▼
Ngân hàng (tài khoản driver — ngoài phạm vi kiểm soát của Panda)
```

## 2.2 Mô tả từng bước — tiền ở đâu, bao lâu, ai chịu trách nhiệm

| Bước | Tiền nằm ở đâu | Thời gian giữ | Ai chịu trách nhiệm |
|---|---|---|---|
| **Khách → Payment Gateway** | Bên trong hệ thống ngân hàng/thẻ của khách, sau đó chuyển vào tài khoản trung gian của Payment Gateway | Tức thời đến vài giây (xử lý giao dịch) | Ngân hàng phát hành thẻ + Payment Gateway (chịu trách nhiệm PCI-DSS, không phải Panda) |
| **Payment Gateway → Panda** | Tài khoản trung gian của Payment Gateway | Theo chu kỳ đối soát của đối tác (thường T+1 đến T+2, ngoài kiểm soát của Panda) | Payment Gateway (đối tác bên thứ ba) |
| **Panda → Ví tạm giữ** | Tài khoản merchant của Panda, nhưng được đánh dấu **Locked** trong ledger nội bộ ngay khi rider xác nhận đặt xe (không đợi đến lúc trả tiền — vì với thanh toán ví Panda, tiền đã nằm sẵn trong Passenger Wallet và bị khoá tại thời điểm xác nhận, đúng theo BRB §5.10) | Từ lúc xác nhận đặt xe đến lúc Trip Finished + Settlement hoàn tất — thường vài phút đến vài giờ, **không bao giờ quá 1 chu kỳ Settlement Queue** (PHẦN 3.9) | Panda (Settlement Engine — PHẦN 4) |
| **Ví tạm giữ → Commission/Voucher/Khuyến mãi/Thuế** | Phân bổ nội bộ trong ledger Panda — không phải dòng tiền vật lý mới ra ngân hàng, mà là **phép chia kế toán** áp dụng đúng công thức BRB §6.3/§6.4/§6.5 | Tức thời tại thời điểm Settlement | Settlement Engine (PHẦN 4), giám sát bởi CFO qua Financial KPI (PHẦN 10) |
| **→ Driver Wallet (Pending)** | Ledger nội bộ Panda, gắn với driver cụ thể, trạng thái **Pending** | Đến chu kỳ payout tiếp theo (weekly mặc định, BRB §6.9) | Panda |
| **Driver Wallet (Pending) → Available** | Chuyển trạng thái tại mốc payout cycle | Tức thời tại mốc chu kỳ | Settlement Engine |
| **Driver Wallet (Available) → Rút tiền** | Rời khỏi Driver Wallet, đi vào hàng đợi xử lý chuyển khoản ngân hàng | T+1 ngày làm việc (BRB §5.5, áp dụng tương tự cho driver) | Panda (bộ phận Tài chính/Payout) |
| **Rút tiền → Ngân hàng driver** | Hoàn toàn ngoài Panda | — | Ngân hàng driver; Panda chỉ còn trách nhiệm đối soát (BRB §5.13) |

## 2.3 Nguyên tắc bất biến của Money Flow

1. **Không đồng nào được coi là doanh thu Panda cho đến khi Settlement chạy xong.** Trước đó, mọi thứ là "Locked Balance" — nợ tiềm năng với rider (nếu huỷ chuyến) hoặc với driver (nếu chuyến hoàn tất).
2. **Thuế không bao giờ đi qua Platform Wallet.** Tách vào Tax Holding ngay tại bước Settlement, tránh rủi ro nhầm lẫn doanh thu thật với tiền giữ hộ nhà nước.
3. **Khuyến mãi dạng "giảm giá tại thời điểm đặt" (voucher) không tạo dòng tiền mới** — nó chỉ làm giảm số tiền rider phải trả. Khuyến mãi dạng "hoàn tiền sau" (cashback) **tạo dòng tiền thật** từ Promotion Fund vào Passenger Wallet. Hai loại này phải được ghi sổ khác nhau (đã phân biệt tại BRB §5.7).

---

# PHẦN 3 — WALLET ARCHITECTURE

## 3.1 Nguyên tắc chung

Mọi ví trong hệ thống là **ledger** (sổ cái cộng dồn giao dịch bất biến), không phải một con số số dư có thể ghi đè — kế thừa nguyên văn BRB §5.2 và áp dụng đồng nhất cho **tất cả** loại ví dưới đây, không chỉ Rider Wallet như BRB hiện mô tả.

## 3.2 Passenger Wallet

**Vai trò:** phương thức thanh toán trả trước (BRB §5.1) — không phải tài khoản ngân hàng, không sinh lãi.
**Số dư con:** Available / Pending / Frozen / Locked / Expired — đã định nghĩa đầy đủ tại BRB §5.8–§5.11. Tài liệu này không định nghĩa lại, chỉ tái sử dụng.

## 3.3 Driver Wallet

**Vai trò:** nơi ghi nhận thu nhập driver trước khi chuyển thành tiền rút được.
**Số dư con [MỚI]** — mở rộng đồng dạng với Passenger Wallet (BRB chưa mô tả tường minh cấu trúc trạng thái cho Driver Wallet, chỉ mô tả lịch payout tại §6.9):
- **Pending:** thu nhập đã ghi nhận từ chuyến hoàn tất, chưa đến chu kỳ payout.
- **Available:** đã qua mốc chu kỳ payout, sẵn sàng rút.
- **Frozen:** đang bị điều tra gian lận (áp dụng cùng quy tắc tối đa 30 ngày như BRB §5.9).

## 3.4 Platform Wallet **[MỚI]**

**Vai trò:** nơi Panda ghi nhận doanh thu đã thực nhận (Commission + Booking Fee, sau khi trừ Promotion/Voucher đã áp dụng ở BRB §6.5 Case A) — cụ thể hoá khái niệm "owned by FAIRRIDE from the moment of trip completion" đã có tại BRB §6.2, đặt tên thành một ledger tường minh để Settlement Engine (PHẦN 4) có nơi ghi vào.
**Không phải** nơi giữ tiền thuế (xem Tax Holding) hay ngân sách khuyến mãi (xem Promotion Fund) — Platform Wallet chỉ chứa phần Panda **thực sự sở hữu**.

## 3.5 Promotion Fund **[MỚI]**

**Vai trò:** cụ thể hoá kỷ luật ngân sách đã có tại BRB §3.3 (Campaign Budget Rules) thành một ledger có số dư thật — mỗi campaign được cấp một hạn mức từ quỹ này; khi quỹ cạn, campaign tự động tạm dừng (đúng quy tắc đã có, không phải quy tắc mới).
**Nguồn nạp:** phân bổ ngân sách quý/tháng do CFO/CPO phê duyệt (BRB §3.3 Rule 1).
**Nguồn rút:** mỗi voucher/promotion do Panda tài trợ áp dụng thành công.

## 3.6 Insurance Fund **[MỚI]**

**Vai trò:** ví dự phòng cho ngày sản phẩm bảo hiểm được ký kết (hiện BRB Appendix B UBD-004 xác nhận **chưa có đối tác bảo hiểm nào** — quỹ này hiện tại **rỗng theo thiết kế**, không phải bị bỏ quên). PSR §"ASSUMPTION" đã gắn cờ `AssumedInsuranceCostVND = 0` khớp chính xác với trạng thái này.
**Kích hoạt:** chỉ bắt đầu nhận dòng tiền khi UBD-004 được đóng và một tỷ lệ trích lập trên mỗi chuyến được CFO phê duyệt chính thức.

## 3.7 Tax Holding **[MỚI]**

**Vai trò:** giữ phần VAT/thuế thu hộ nhà nước (PSR đã gắn cờ `AssumedVATRate = 10%` là ASSUMPTION, chưa được CFO xác nhận chính thức — xem đề xuất bổ sung BRB cuối tài liệu). Về nguyên tắc thiết kế: đây **không phải tiền của Panda**, không được dùng cho bất kỳ mục đích vận hành nào, chỉ chờ kỳ nộp thuế định kỳ.

## 3.8 Pending Balance / Frozen Balance / Available Balance

Áp dụng thống nhất cho **mọi** loại ví ở trên (không riêng Passenger Wallet như BRB hiện mô tả) — đây là phần mở rộng **[MỚI]** duy nhất về phạm vi áp dụng, không phải khái niệm mới (bản thân 3 trạng thái này đã có tại BRB §5.8–§5.11).

## 3.9 Settlement Queue **[MỚI]**

**Vai trò:** hàng đợi xử lý các chuyến đã "Trip Finished" nhưng chưa chạy xong Settlement Engine (PHẦN 4). Cần thiết vì Settlement không phải là một thao tác tức thời an toàn để chạy đồng thời hàng loạt (rủi ro race-condition khi nhiều chuyến của cùng một driver settle cùng lúc) — nguyên tắc xử lý tuần tự theo driver, song song giữa các driver khác nhau, cùng triết lý với cơ chế **idempotency key + saga compensation** đã áp dụng thật cho Booking Service (MVP §2.1, Hardening H3–H4) — Settlement Engine tái sử dụng đúng pattern kiến trúc này ở tầng nghiệp vụ, không phát minh cơ chế mới.

---

# PHẦN 4 — SETTLEMENT ENGINE

## 4.1 Quy trình

```
Trip Finished (driver bấm "Hoàn thành" hoặc hệ thống tự động khi rider xác nhận điểm đến)
   ▼
Payment (thu tiền thật từ Payment Gateway/Passenger Wallet — hoặc xác nhận số dư khoá đã đủ)
   ▼
Settlement (tính Commission/Voucher/Khuyến mãi/Thuế — PHẦN 2.2, đưa vào Settlement Queue nếu hệ thống đang bận)
   ▼
Wallet Update (ghi nhận song song: Platform Wallet +Commission, Tax Holding +VAT, Driver Wallet(Pending) +Net Driver)
   ▼
Commission (đã tính ở bước Settlement, ghi sổ tại đây)
   ▼
Tax (đã tính ở bước Settlement, ghi sổ tại đây)
   ▼
Driver Income (số dư Pending của driver tăng — driver nhìn thấy trong app ngay, nhưng chưa rút được)
   ▼
Withdrawal (chờ đến chu kỳ payout hoặc driver dùng Fast Withdrawal nếu đủ điều kiện tier)
```

## 4.2 Thời gian xử lý

| Bước | Thời gian mục tiêu |
|---|---|
| Trip Finished → Payment | Tức thời (< vài giây) nếu thanh toán bằng Passenger Wallet (tiền đã khoá sẵn); phụ thuộc Payment Gateway nếu thanh toán bằng thẻ |
| Payment → Settlement | Tức thời nếu Settlement Queue rảnh; tối đa vài phút nếu đang xử lý dồn (giờ cao điểm) |
| Settlement → Wallet Update | Tức thời (một giao dịch nguyên tử — atomic — theo đúng nguyên tắc "atomic transactions" đã áp dụng thật cho Booking/Dispatch, MVP §2.1) |
| Driver Wallet (Pending) → Available | Theo chu kỳ payout: **hàng tuần** mặc định (BRB §6.9), hoặc **hàng ngày** nếu driver đủ điều kiện Fast Withdrawal (tier Gold+, có phí) |
| Available → Ngân hàng | T+1 ngày làm việc (BRB §5.5, áp dụng tương tự) |

## 4.3 Điều kiện để Settlement chạy

- Trip phải ở trạng thái `completed` hoặc `settled` (không chạy Settlement cho trip `cancelled` — đi qua nhánh Refund riêng, BRB §6.7).
- Payment phải xác nhận thành công (nếu thanh toán thẻ thất bại, Settlement **không chạy**, trip chuyển vào trạng thái chờ thanh toán, không tự động ghi nhận thu nhập driver cho đến khi thanh toán thành công — driver vẫn có Minimum Fare Guarantee theo BRB §2.14 bất kể, đảm bảo driver không bị treo thu nhập vì lỗi thanh toán của rider).
- Không có cờ gian lận đang mở (Risk Score, BRB §12.1) trên trip đó — nếu có, Settlement tạm hoãn, trip vào diện Manual Review (BRB §11.11).

## 4.4 Rollback

Áp dụng đúng theo BRB §6.7/§6.8 (Refund Settlement, Chargeback):
- **Lỗi phía driver** (BRB §6.7 "Driver-fault refund"): thu nhập driver đã ghi Pending bị **đảo ngược bằng bút toán bù trừ** (offsetting entry, đúng nguyên tắc bất biến ledger tại BRB §5.12/§6.7 — không xoá, không sửa) — nếu đã chuyển sang Available nhưng chưa rút, trừ trực tiếp; nếu đã rút, ghi nợ chờ trừ vào chu kỳ tiếp theo (tối đa 10% mỗi kỳ theo đúng BRB §8.13B để không gây khó khăn tài chính đột ngột).
- **Lỗi phía platform** (BRB §6.7 "Platform-fault refund"): driver **giữ nguyên** thu nhập đã ghi nhận; Panda chịu toàn bộ chi phí hoàn tiền, ghi nhận là tổn thất vận hành riêng (không trừ vào Commission Revenue — để không làm sai lệch KPI, xem PHẦN 10).
- **Chargeback** (BRB §6.8): Settlement Engine đóng băng (Frozen) phần thu nhập driver liên quan cho đến khi điều tra nội bộ kết thúc; nếu chargeback gian lận từ phía rider, driver được đền bù đầy đủ, Panda chịu rủi ro (BRB §6.8 "Driver protection").

## 4.5 Retry

Khi một bước trong chuỗi thất bại (ví dụ: Wallet Update timeout do lỗi hạ tầng), Settlement Engine áp dụng cơ chế **idempotency key** giống hệt Booking Service (MVP §2.1) — retry an toàn không tạo double-settlement, vì mỗi lần chạy lại dùng cùng một khoá idempotent gắn với `trip_id`, không tạo bút toán trùng. Số lần retry tối đa và backoff cụ thể là chi tiết kỹ thuật, nằm ngoài phạm vi tài liệu business-level này — chỉ nguyên tắc "idempotent, không tạo tiền trùng lặp" là bất biến ở tầng nghiệp vụ.

---

# PHẦN 5 — COMMISSION ENGINE

## 5.1 Framework, không phải bảng số

BRB §7.1 đã định nghĩa 5 bậc (Bronze/Silver/Gold/Platinum/Diamond) với tỷ lệ cụ thể — tài liệu này **không lặp lại con số** (đúng yêu cầu sprint), chỉ mô tả **cách engine hoạt động**:

1. **Đầu vào:** `driver_tier` (kết quả từ Driver Performance Engine, BRB Part 9 — Acceptance Rate/Completion Rate/Rating/Trip Volume/Trust Score).
2. **Tra cứu:** Commission Engine tra một **bảng cấu hình** (không hardcode trong logic tính giá — xem PHẦN 11) ánh xạ `tier → commission_rate`.
3. **Áp dụng:** tỷ lệ được nhân với `commission_base` (Base+Distance+Time+surcharge+Airport+Waiting — đã định nghĩa chính xác tại BRB §2.2.6/§2.2.7/§2.2.9, không đổi ở đây).
4. **Không áp dụng lên:** Booking Fee (100% Panda), Toll/Bridge/Parking (100% driver pass-through) — quy tắc này bất biến, độc lập với tier.

## 5.2 Vì sao phải qua Rule Engine, không hardcode tỷ lệ

Ba lý do, cả ba đều đã được minh chứng thực nghiệm qua PSR PHẦN 5 (Optimization): (1) tỷ lệ commission là đòn bẩy **nhạy nhất** với vị thế cạnh tranh (PSR cho thấy hạ 4 điểm % cải thiện tỷ lệ driver-thắng-đối-thủ từ 50.7%→63.7%) — cần thay đổi được **mà không phải sửa code**; (2) BRB §1.4 yêu cầu "giảm hoa hồng phải báo trước 30 ngày" — một Rule Engine có **hiệu lực theo ngày** (effective-dated rule) là cách duy nhất để tự động hoá cam kết này mà không phụ thuộc vào một kỹ sư nhớ deploy đúng ngày; (3) tỷ lệ có thể khác nhau theo thành phố (BRB §2.18 Multi-city) — cấu hình theo bảng cho phép mỗi thị trường có bảng riêng mà không nhân bản logic.

## 5.3 Khả năng thay đổi qua Rule Engine

Commission Engine phải hỗ trợ (mô tả nghiệp vụ, không phải đặc tả kỹ thuật):
- Thay đổi tỷ lệ theo tier mà **không cần release ứng dụng**.
- Đặt **ngày hiệu lực trong tương lai** (để tuân thủ quy tắc báo trước 30 ngày).
- Áp dụng theo **thành phố/thị trường** khác nhau.
- Giữ **lịch sử đầy đủ** mọi thay đổi tỷ lệ (ai đổi, khi nào, từ bao nhiêu thành bao nhiêu) — phục vụ audit và đúng nguyên tắc minh bạch (PHẦN 1.2).

---

# PHẦN 6 — PROMOTION ENGINE

## 6.1 Danh mục (đối chiếu nguồn)

| Loại | Nguồn | Trạng thái |
|---|---|---|
| Voucher | BRB Part 4 | Đã có luật đầy đủ |
| Campaign (Coupon) | BRB §3.2.9 | Đã có |
| Referral | BRB §3.2.7 | Đã có |
| First Ride | BRB §3.2.1 | Đã có |
| Birthday | BRB §3.2.2 | Đã có |
| Weekend | BRB §3.2.4 | Đã có |
| Comeback (Win-back) | PS §7.2.4 | **[MỚI]** |
| Membership | PS §7.2.1 | **[MỚI]** |
| Student | PS §7.2.2 | **[MỚI]** |
| Airport | — | **[MỚI]** — chưa có promotion riêng cho sân bay (chỉ có Airport Fee là phụ phí, BRB §2.2.7 — chưa có khuyến mãi bù/thu hút cho phân khúc sân bay) |
| Night | Gần với Rain Campaign (BRB §3.2.5) về cơ chế | **[MỚI]** — chưa có promotion riêng bù Night Surcharge, chỉ mới có surcharge |
| Referral Reward | BRB §3.2.7 | Đã có (một phần của Referral) |

## 6.2 Điều kiện, ưu tiên, xung đột

**Không định nghĩa lại** — toàn bộ khung xung đột/ưu tiên/stacking đã có sẵn đầy đủ tại BRB §3.4 (Campaign Priority Rules — thứ tự Voucher > Referral > First Ride > Festival > Golden Hour/Rain/Weekend > Birthday > Cashback), §3.5 (Campaign Conflict Resolution), §3.6 (Expiration), §3.7 (Quota).

**Nguyên tắc mở rộng cho các loại [MỚI]** (Comeback/Membership/Student/Airport/Night) — phải được **chèn vào đúng vị trí trong bảng ưu tiên BRB §3.4 hiện có**, không tạo bảng ưu tiên song song:
- **Membership** giảm Booking Fee cố định (PS §7.2.1) → không xung đột với voucher giảm % (khác trường tác động), có thể **stack** với voucher (ngoại lệ cần CPO phê duyệt tường minh theo đúng cơ chế "Override" đã có ở BRB §3.4).
- **Student** và **Comeback** là giảm giá theo %/VND như voucher → nằm cùng nhóm ưu tiên với Golden Hour/Rain/Weekend, "mức giảm cao nhất thắng" nếu trùng thời điểm (đúng BRB §3.5).
- **Airport/Night** (nếu triển khai dạng promotion, không chỉ surcharge) nên xếp cùng nhóm Festival — kích hoạt theo điều kiện khách quan (vị trí/giờ), không do rider chủ động chọn.

## 6.3 Nguyên tắc ngân sách

Không thay đổi — mọi promotion [MỚI] ở trên phải tuân thủ nguyên vẹn BRB §3.3 (ngân sách được duyệt trước, tự dừng khi cạn, cảnh báo ở 90%).

---

# PHẦN 7 — DRIVER INCENTIVE

## 7.1 Danh mục (đối chiếu nguồn)

| Loại | Nguồn | Trạng thái |
|---|---|---|
| Quest (Daily/Weekly/Monthly) | BRB §8.2–§8.4 | Đã có |
| Guaranteed Income | BRB §8.9 | Đã có |
| Airport Bonus | BRB §8.7 | Đã có |
| Rain Bonus | BRB §8.8 | Đã có |
| Peak Bonus | BRB §8.6 (Peak Hour Bonus) | Đã có |
| Referral (driver) | BRB §8.13 | Đã có |
| Weekly Mission | BRB §8.3 | Đã có |
| Monthly Mission | BRB §8.4 (Monthly Incentive) | Đã có |
| **Heat Bonus** | — | **[MỚI]** — chưa có trong BRB Part 8; liên hệ trực tiếp với Heat Map (PS §5.2) |
| **Night Bonus** | — | **[MỚI]** — BRB chỉ có Night Surcharge (phụ phí cước, tự động chia theo commission), chưa có thưởng cố định riêng cho ca đêm như Airport/Rain Bonus |

## 7.2 Engine — mô tả framework (không ghi số tiền)

Mọi loại thưởng ở trên vận hành theo cùng một khung 4 bước:
1. **Điều kiện kích hoạt** (trigger): mốc số chuyến (Quest), vùng nhiệt độ cầu cao (Heat Bonus), khung giờ (Peak/Night), khu vực (Airport), thời tiết (Rain), hoặc sự kiện (Referral driver hoàn thành 50 chuyến đầu).
2. **Đối tượng đủ điều kiện**: lọc theo Trust Score/Anti-Abuse (PHẦN 9) — không trả thưởng cho tài khoản đang bị điều tra (BRB §8.14 Rule 2).
3. **Tính toán**: tra bảng cấu hình theo Rule Engine (PHẦN 11) — số tiền là tham số, không phải hằng số trong logic.
4. **Ghi nhận**: cộng vào Driver Wallet (Pending) tại mốc BRB §8.13B (06:00 sáng hôm sau), có dòng riêng biệt trong sổ thu nhập, có cửa sổ khiếu nại 7 ngày.

## 7.3 Heat Bonus — mô tả khái niệm **[MỚI]**

Thưởng theo thời gian thực khi driver hoạt động trong một "vùng nhiệt" (zone có DSR cao nhưng **chưa đạt ngưỡng kích hoạt Surge**, BRB §2.13.2 — tức là vùng đang nóng lên nhưng chưa đủ nóng để tính surge cho rider). Mục đích: **kéo driver đến vùng thiếu cung TRƯỚC KHI** rider phải trả giá surge — nếu Heat Bonus hoạt động hiệu quả, số lần thực sự phải kích hoạt Surge sẽ **giảm** (thanh khoản tốt hơn = ít cần công cụ giá để cân bằng cung-cầu hơn). Nguồn tài trợ: Platform (không phải rider) — khác biệt căn bản với Surge, vốn do rider trả.

---

# PHẦN 8 — MEMBERSHIP

## 8.1 Nguyên tắc bất biến (áp dụng cho cả hai phía)

**Membership không bao giờ thay đổi công thức giá cước** — chỉ thay đổi **dịch vụ đi kèm**. Đây là ranh giới cứng để không vi phạm Constitution Article II §2.1 (Rider Fairness: hai rider cùng chuyến, cùng điều kiện, cùng giá) — nguyên tắc này áp dụng chặt cho Passenger Membership; với Driver Membership (tức là Driver Tier, đã tồn tại tại BRB §7.1), commission **có** khác theo tier — điều này **không** vi phạm nguyên tắc công bằng vì nó thưởng cho **hiệu suất đã chứng minh** (số chuyến, tỷ lệ nhận/hoàn thành, đánh giá), không phải một đặc quyền mua được bằng tiền.

## 8.2 Passenger Membership — Free / Silver / Gold / Diamond **[MỚI toàn bộ]**

BRB §10.5 hiện ghi rõ: "Loyalty tiers are not implemented at launch." PS §7.2.1 mới chỉ phác thảo một hạng "Panda Plus" duy nhất. Tài liệu này đề xuất mở rộng thành 4 hạng, thiết kế mới hoàn toàn:

| Hạng | Điều kiện nâng hạng (dựa trên hành vi, không mua được trực tiếp bằng tiền — trừ Diamond) | Quyền lợi (dịch vụ, không phải giá) |
|---|---|---|
| Free | Mặc định | Không có quyền lợi đặc biệt |
| Silver | Số chuyến hoàn thành trong 90 ngày vượt ngưỡng cấu hình (Rule Engine, PHẦN 11) | Ưu tiên hỗ trợ nhanh hơn mức chuẩn |
| Gold | Ngưỡng chuyến cao hơn + không có vi phạm (huỷ quá mức, BRB §10.1 Abuse) | Ưu tiên ghép chuyến nhẹ (tương tự Priority Dispatch của driver, BRB §7.5, nhưng biên độ ưu tiên nhỏ hơn để không ảnh hưởng ETA của rider khác quá nhiều) |
| Diamond | Ngưỡng chuyến cao nhất, **hoặc** mua gói thuê bao (mô hình PS §7.2.1 Panda Plus) | Toàn bộ quyền lợi Gold + trần surge cá nhân thấp hơn trần chung (ví dụ ×1.5 thay vì ×2.0 — đây là **giới hạn giá cao nhất họ phải trả trong điều kiện surge**, không phải giá thấp hơn ở điều kiện thường, nên không vi phạm nguyên tắc "cùng giá" cho chuyến **không surge**) |

**Lưu ý về ngoại lệ trần surge:** đây là ranh giới tế nhị — về mặt kỹ thuật nó tạo ra hai rider trả giá khác nhau cho *cùng một chuyến surge*. Tài liệu này xếp đây vào nhóm "cần CPO+Legal xác nhận rõ ràng có vi phạm tinh thần Rider Fairness hay không" trước khi triển khai — **không tự động coi là đã được duyệt** chỉ vì xuất hiện trong tài liệu thiết kế này.

## 8.3 Driver Membership — Bronze / Silver / Gold / Diamond / Elite **[MỚI: chỉ riêng "Elite"]**

Bronze→Diamond đã có đầy đủ tại BRB §7.1–§7.6 (điều kiện, quyền lợi, đánh giá lại hàng tháng, downgrade có cảnh báo trước). Tài liệu này đề xuất thêm hạng **Elite** phía trên Diamond:

**Điều kiện nâng hạng đề xuất** (theo đúng khuôn mẫu leo thang của BRB §7.2, không sáng tạo tiêu chí mới): ngưỡng chuyến lifetime cao hơn Diamond, Acceptance/Completion/Rating cao hơn Diamond, **cộng thêm** điều kiện định tính mà BRB chưa có ở bất kỳ tier nào: **zero khiếu nại nghiêm trọng trong 360 ngày VÀ đã hoàn thành ít nhất 1 chương trình đào tạo/an toàn của Panda** (liên hệ Safety Center đã xây ở sprint UI trước — tạo lý do nghiệp vụ thực sự cho tính năng đó, không chỉ là UI).

**Quyền lợi đề xuất:** toàn bộ quyền lợi Diamond + được mời vào hội đồng phản hồi sản phẩm (Driver Advisory — phi tài chính, tăng gắn kết) + ưu tiên tuyệt đối trong Priority Dispatch (trên cả Diamond thường).

---

# PHẦN 9 — ANTI ABUSE

## 9.1 Đối chiếu nguồn — hầu hết đã có luật, tài liệu này chỉ tổ chức lại theo góc nhìn "dòng tiền"

| Loại | Nguồn BRB | Giải pháp (mô tả nghiệp vụ) |
|---|---|---|
| Voucher Fraud | §11.3 | Giới hạn 1 voucher/tài khoản đã xác minh (BRB §4.15), mã voucher ngẫu nhiên hoá mật mã, cờ nếu >3 voucher/7 ngày trên chuyến <2km |
| Referral Fraud | §11.4 | Người giới thiệu/được giới thiệu phải khác thiết bị, khác phương thức thanh toán; thưởng giữ 7 ngày trước khi trả (BRB §3.2.7) |
| Fake GPS | §11.1 (GPS Spoofing) | Route Engine riêng của Panda (không phụ thuộc bên thứ 3) đối chiếu tốc độ/quỹ đạo hợp lý |
| Fake Trip | §11.2 | Yêu cầu chuyến ≥1km, ≥3 phút, chuyển động thật mới được tính vào bất kỳ Incentive nào (BRB §8.14 Rule 1) |
| Driver Farming | §11.6 (Multi-Account) + §11.7 (Device Farming) | Một thiết bị vật lý không được vận hành nhiều tài khoản driver đồng thời; địa chỉ nhà/IP lặp lại giữa nhiều tài khoản mới bị gắn cờ mềm |
| Passenger Farming | §11.6 + §11.7 (áp dụng tương tự cho rider) | Cùng cơ chế Device Farming, áp dụng phía rider — hiện BRB chỉ mô tả tường minh cho driver, đề xuất mở rộng tường minh sang rider (xem đề xuất bổ sung BRB cuối tài liệu) |
| Collusion | §11.5 | Driver và rider "quen nhau" tạo chuyến giả để farm Incentive/Voucher — phát hiện qua mẫu hình lặp lại tuyến đường/cặp tài khoản |
| **Self Ride** | Là một trường hợp riêng của Collusion (§11.5), chưa được đặt tên tường minh trong BRB | Driver dùng chính tài khoản rider của mình (hoặc người thân cùng hộ khẩu/cùng thiết bị) để tự đặt chuyến — phát hiện qua trùng khớp thiết bị/địa chỉ thanh toán giữa driver_id và rider_id của cùng một chuyến |
| Chargeback | §11.9 (Payment Fraud) + §6.8 | Điều tra nội bộ trước khi hoàn/giữ; driver luôn được bảo vệ khỏi chargeback gian lận (BRB §6.8) |

## 9.2 Nguyên tắc thiết kế chung cho Anti-Abuse trong Economy Engine

Mọi cơ chế chống gian lận ở trên phải **chặn trước khi tiền rời khỏi Ví tạm giữ** (PHẦN 3.9), không chặn sau khi đã vào Driver Wallet Available — vì thu hồi tiền đã rút (PHẦN 4.4 Rollback) luôn tốn kém và ảnh hưởng trải nghiệm hơn giữ lại đúng lúc. Đây là lý do kiến trúc thực sự của Settlement Queue (PHẦN 3.9): nó là **điểm chặn cuối cùng** trước khi tiền trở thành "sự thật không thể đảo ngược dễ dàng".

---

# PHẦN 10 — FINANCIAL KPI

## 10.1 Đã có định nghĩa tại BRB §14.1 — không định nghĩa lại

GMV, Net Revenue, Take Rate, Driver Earnings, Promotion Cost, Voucher Cost, Wallet Liability, Outstanding Settlement.

## 10.2 Bổ sung **[MỚI]** — chưa có trong BRB §14.1

| KPI | Định nghĩa |
|---|---|
| **Contribution Margin** | Net Revenue trừ đi toàn bộ chi phí biến đổi trực tiếp trên mỗi chuyến (Promotion Cost + Minimum Earning Guarantee top-up + Long Pickup Compensation nếu được duyệt — PSR đã định lượng các khoản này là có thật, không phải lý thuyết) — chỉ số cho biết "mỗi chuyến thêm có thật sự đóng góp dương hay không", tách biệt khỏi chi phí cố định vận hành công ty |
| **Burn Rate** | Tốc độ tiêu ngân sách tăng trưởng (Promotion Fund + Incentive Fund) trên một đơn vị thời gian — đo bằng VND/tháng, theo dõi song song với Contribution Margin để đảm bảo tăng trưởng không kéo dài vô thời hạn ở mức lỗ ròng |
| **Commission Revenue** | Riêng phần Commission (tách khỏi Booking Fee) trong Net Revenue — cần thiết vì PHẦN 5 cho thấy Commission là đòn bẩy chính sách chính, cần theo dõi độc lập với Booking Fee (vốn ổn định hơn) |
| **Average Fare** | Customer Total trung bình trên mỗi chuyến hoàn tất — chỉ số sức khoẻ thị trường cơ bản nhất |
| **Average Incentive** | Tổng chi Driver Incentive (PHẦN 7) chia cho số chuyến — theo dõi để đảm bảo không vượt ngân sách đã duyệt theo cùng kỷ luật ngân sách như Promotion (BRB §3.3) |
| **CAC** (Customer Acquisition Cost) | Tổng chi phí thu hút một Rider mới (Promotion Cost dành riêng cho First Ride/Referral + chi phí marketing ngoài phạm vi tài liệu này) chia cho số Rider mới có ít nhất 1 chuyến hoàn tất |
| **LTV** (Lifetime Value) | Tổng Net Revenue kỳ vọng từ một Rider trong toàn bộ vòng đời sử dụng — dùng làm mẫu số so sánh với CAC (nguyên tắc: CAC phải nhỏ hơn LTV đáng kể mới là tăng trưởng lành mạnh, đúng tinh thần BRB §3.1 "lifetime value của rider biện minh cho CPA cao") |
| **Payback Period** | Thời gian (số tháng) để Net Revenue tích luỹ từ một Rider bù lại CAC đã chi cho họ |

## 10.3 Vai trò của KPI trong Economy Engine

Không phải để báo cáo — mà để **là điều kiện dừng tự động** cho Rule Engine (PHẦN 11): nếu Burn Rate vượt ngưỡng đã duyệt, hoặc Contribution Margin âm liên tục quá X chu kỳ báo cáo, hệ thống phải tự động cảnh báo CFO theo đúng cơ chế đã có tại BRB §14.6 ("Breach response... trình bày cho CEO trong 15 ngày làm việc") — không có KPI nào trong tài liệu này được thu thập chỉ để trưng bày, mọi KPI đều gắn với một hành động khi vượt ngưỡng.

---

# PHẦN 11 — RULE ENGINE

## 11.1 Nguyên tắc "không hardcode" áp dụng cho toàn bộ Economy Engine

Pricing, Commission, Promotion, Membership, Quest, Heat Bonus, Voucher — **tất cả** phải là **dữ liệu cấu hình**, không phải logic cứng trong code. Đây không phải là một nguyên tắc trừu tượng — nó đã được **chứng minh khả thi** ngay trong chính sprint mô phỏng giá (PSR): toàn bộ hằng số (base fare, commission theo tier, surcharge multiplier, surge band...) đã được tách hoàn toàn khỏi logic tính toán vào một nơi duy nhất, mỗi giá trị có thể đổi mà không đổi công thức. Rule Engine ở cấp toàn nền tảng là việc **tổng quát hoá pattern đã chứng minh này** sang mọi engine khác (Commission/Promotion/Membership/Incentive), không phải một ý tưởng chưa kiểm chứng.

## 11.2 Đặc tính bắt buộc của Rule Engine (mô tả nghiệp vụ)

1. **Versioned:** mọi thay đổi rule tạo ra một phiên bản mới, không ghi đè phiên bản cũ (cùng triết lý bất biến ledger, PHẦN 3.1).
2. **Effective-dated:** rule có thể được duyệt hôm nay, có hiệu lực trong tương lai (bắt buộc để tuân thủ "báo trước 30 ngày" của BRB §1.4).
3. **Scoped:** theo thành phố/thị trường (BRB §2.18), theo loại xe, theo tier — không phải một giá trị toàn cục duy nhất.
4. **Audit trail đầy đủ:** ai duyệt, khi nào, giá trị cũ/mới — phục vụ trực tiếp nguyên tắc minh bạch (PHẦN 1.2) và quy trình tu chính tài liệu (Constitution Article XI, áp dụng tương tự cho thay đổi rule vận hành).
5. **Kiểm tra an toàn trước khi áp dụng:** mọi thay đổi rule ảnh hưởng đến tiền phải được **mô phỏng trước** bằng Simulation Engine (đã xây tại sprint trước, `backend/services/pricing/simulation`) trước khi áp dụng vào production — đây là quy trình gate bắt buộc, không tuỳ chọn: "Chỉ khi simulation được chứng minh tốt mới đề xuất merge" (nguyên văn yêu cầu sprint Simulation Engine) áp dụng cho **mọi** thay đổi rule về tiền, không riêng gì Pricing.

## 11.3 Ai được sửa Rule Engine

Theo đúng mô hình phân quyền quyết định đã có tại Constitution Article V (Class 2/3 decisions) — thay đổi Commission/Promotion ngân sách lớn là quyết định Class 2 (CPO+CFO), thay đổi tham số một campaign nhỏ trong ngân sách đã duyệt là Class 3 (Tech Lead vận hành/Marketing Ops) — tài liệu này không tạo mô hình phân quyền mới, chỉ ánh xạ Rule Engine vào mô hình đã có.

---

# PHẦN 12 — ROADMAP

Khác với PS §4 (giai đoạn theo **số lượng driver**, dùng cho chiến lược giá), roadmap này chia theo **số lượng rider** (đúng yêu cầu sprint) — hai roadmap bổ trợ nhau, không mâu thuẫn: tỷ lệ rider/driver lành mạnh trong PS đã ngụ ý một rider tương ứng khoảng 5–10 chuyến/tháng, nên mốc driver trong PS và mốc rider ở đây tăng theo cùng nhịp.

| Giai đoạn | Module BẬT | Module CHƯA bật (có chủ đích) |
|---|---|---|
| **Launch** (0 rider — MVP, đúng theo `MVP_DEVELOPMENT_PLAN.md` hiện trạng) | Money Flow cơ bản (Khách→Gateway→Panda→Driver, chưa cần Escrow phức tạp vì volume thấp, rủi ro thấp); Commission Engine (bảng tĩnh, chưa cần Rule Engine động); Settlement Engine (chu kỳ payout đơn giản, thủ công chấp nhận được ở quy mô này) | Promotion Engine đầy đủ (chỉ First Ride); Membership (chưa ai đủ điều kiện nâng hạng); Heat Bonus (cần đủ dữ liệu zone); Anti-Abuse nâng cao (Collusion detection cần đủ mẫu hình lịch sử) |
| **10.000 rider** | Promotion Fund + Voucher Engine đầy đủ (đủ volume để campaign có ý nghĩa thống kê); Driver Incentive cơ bản (Quest/Guaranteed Income); Tax Holding chính thức (đủ doanh thu để việc tách bạch thuế trở nên quan trọng về tuân thủ) | Membership 4 hạng (cần lịch sử hành vi đủ dài để phân hạng công bằng); Elite driver tier (chưa ai đạt ngưỡng lifetime) |
| **100.000 rider** | Heat Bonus (đủ mật độ dữ liệu zone để tính "vùng nóng" có ý nghĩa); Passenger Membership Silver/Gold; Rule Engine hoá toàn bộ (không còn chấp nhận bảng tĩnh thủ công ở quy mô này — rủi ro vận hành thủ công vượt ngưỡng an toàn); Insurance Fund kích hoạt (nếu UBD-004 đã đóng) | Diamond Membership thuê bao (Panda Plus) — cần hạ tầng thu phí định kỳ ổn định trước |
| **1.000.000 rider** | Passenger Membership Diamond đầy đủ (kể cả trần surge cá nhân — sau khi Legal xác nhận PHẦN 8.2); Elite driver tier; Financial KPI Burn Rate/Contribution Margin giám sát tự động thời gian thực (không còn báo cáo thủ công) | Đa thị trường/đa thành phố cho toàn bộ Economy Engine đồng thời (nên làm tuần tự từng thị trường, không bật đồng loạt) |
| **10.000.000 rider** | Toàn bộ Economy Engine vận hành đa thị trường, đa tiền tệ (BRB §15.9 International Expansion); Anti-Abuse với mô hình phát hiện theo mẫu hình lịch sử quy mô lớn (vẫn rule-based có kiểm soát, không nhất thiết cần ML — đúng tinh thần "logic, không AI" đã áp dụng cho Dynamic Pricing ở PS §5) | — (ở quy mô này, câu hỏi không còn là "bật module gì" mà là "module nào cần tách thành service độc lập vì tải", nằm ngoài phạm vi tài liệu business-level này) |

---

# PHẦN 13 — NHỮNG THỨ KHÔNG NÊN LÀM

| Điều không nên làm | Lý do |
|---|---|
| **Không đốt tiền vô tội vạ** | Đã là nguyên tắc gốc từ PS §8.3 — mọi ngân sách khuyến mãi/incentive phải có điều kiện dừng (BRB §3.3), có đo lường ROI (CPIR, BRB §3.8), không có "quỹ chiến tranh" nằm ngoài kiểm soát tài chính thông thường |
| **Không giảm giá toàn quốc đồng loạt** | Vi phạm trực tiếp nguyên tắc "Rider Fairness" nếu áp dụng không có mục tiêu — giảm giá nên **có mục tiêu** (targeted, theo hành vi/rủi ro rời bỏ — đúng cách PS §8.2 xử lý kịch bản "nếu Grab giảm giá") thay vì đại trà, vừa tốn kém hơn vừa kém hiệu quả hơn (giảm giá cho người đã trung thành = lãng phí ngân sách) |
| **Không thưởng đều cho mọi driver bất kể hiệu suất** | Vi phạm triết lý "công bằng = thưởng cho hiệu suất đã chứng minh" (PHẦN 8.1) — thưởng đều tạo động lực ngược: driver giỏi không có lý do duy trì chất lượng, driver kém không có động lực cải thiện. BRB §8.1 đã cấm rõ: incentive không được "extract maximum working hours... regardless of wellbeing" và không được "incentivise behaviours that harm rider experience" |
| **Không commission cố định vĩnh viễn không qua Rule Engine** | Đã chứng minh tại PHẦN 5.2 — commission là đòn bẩy cạnh tranh cần điều chỉnh theo thời gian thực tế thị trường (PSR), cố định cứng trong code khiến Panda **không thể phản ứng** khi PS §8 yêu cầu (kịch bản đối thủ giảm giá) |
| **Không trợ giá vô hạn** | Đối lập trực tiếp với PHẦN 1.1 — mọi trợ giá (Promotion Fund, Guaranteed Income, Long Pickup Compensation) đều có ngân sách hữu hạn, có kỳ hạn, có điều kiện dừng tự động (Safety Check đã chứng minh khả thi tại PSR PHẦN 7: 0/111 vi phạm an toàn, kể cả trong kịch bản voucher vượt giá trị chuyến) |
| **Không để Wallet là "số dư đơn" có thể ghi đè** | Vi phạm nguyên tắc ledger bất biến (PHẦN 3.1) — một con số ghi đè không thể audit, không phát hiện được lỗi/gian lận sau khi xảy ra |
| **Không settlement tức thời không qua Settlement Queue khi hệ thống đang tải cao** | Rủi ro race-condition đã được cảnh báo tại PHẦN 3.9 — xử lý song song không kiểm soát trên cùng một driver có thể tạo double-payout |
| **Không membership thay đổi công thức giá cước** | Vi phạm Constitution Article II §2.1 — đã giải thích chi tiết tại PHẦN 8.1, đây là ranh giới cứng nhất trong toàn bộ tài liệu |

---

# ĐỀ XUẤT CẦN BỔ SUNG VÀO BUSINESS RULE BIBLE

Để hai tài liệu (BRB và Economy Engine) hợp nhất thành một nguồn duy nhất trong tương lai, các mục sau cần đi qua quy trình tu chính chính thức (Constitution Article XI) trước khi có hiệu lực vận hành:

1. **Đơn vị tiền tệ production đang sai** (không phải quy tắc kinh doanh, nhưng là gap đã phát hiện qua PSR §1.3 cần CPO/CFO xác nhận lại: BRB đã quy định VND, cần đảm bảo mọi tài liệu/hệ thống tương lai không vô tình lặp lại sai lệch này).
2. **VAT rate chính thức** — hiện là ASSUMPTION 10% trong PSR, BRB chưa có mục nào quy định tỷ lệ VAT áp vào công thức cước. Cần CFO xác nhận và bổ sung thành một mục mới trong BRB Part 2 hoặc Part 14.
3. **Cấu trúc trạng thái Driver Wallet** (Pending/Available/Frozen) — hiện chỉ mô tả tường minh cho Rider Wallet (BRB §5.8–§5.11). Đề xuất bổ sung một mục tương đương trong BRB Part 6 (Settlement Engine) hoặc một Part 5B mới (Driver Wallet).
4. **Đặt tên chính thức cho Platform Wallet / Promotion Fund / Insurance Fund / Tax Holding** — hiện các khái niệm sở hữu tiền đã tồn tại rải rác (BRB §6.2 Money Ownership, §3.3 Campaign Budget) nhưng chưa được đặt tên thành các ledger cụ thể. Đề xuất bổ sung một Part mới "Ledger Architecture" thống nhất.
5. **Heat Bonus** — cơ chế thưởng driver theo vùng nhiệt trước khi Surge kích hoạt — cần được thêm vào BRB Part 8 (Driver Incentive Engine) như một mục mới (§8.15 trở đi), có ngân sách và điều kiện chống lạm dụng riêng.
6. **Night Bonus** — thưởng cố định cho ca đêm, tách biệt khỏi Night Surcharge (vốn là phụ phí cước, không phải thưởng) — cần bổ sung vào BRB Part 8.
7. **Passenger Membership 4 hạng (Free/Silver/Gold/Diamond)** — hiện BRB §10.5 mới chỉ nói "chưa triển khai, để giai đoạn sau". Đề xuất đây là nội dung ưu tiên cao nhất cần chính thức hoá, đặc biệt là **câu hỏi trần surge cá nhân cho hạng Diamond** (PHẦN 8.2) cần Legal xác nhận trước khi đưa vào BRB, không chỉ CPO.
8. **Driver Tier "Elite"** — bổ sung vào BRB §7.1–§7.6 như tier thứ 6, với điều kiện định tính (đào tạo an toàn) khác với các tier thuần định lượng hiện có — cần thảo luận liệu BRB có muốn giữ nguyên tắc "tier thuần định lượng" hay chấp nhận thêm tiêu chí định tính.
9. **Mở rộng tường minh Device/Multi-Account Farming (BRB §11.6–§11.7) sang phía Rider**, và đặt tên chính thức pattern "Self Ride" như một mục con của Collusion (§11.5) thay vì chỉ ngụ ý.

---

*Kết thúc tài liệu — Panda Economy Engine — v0.1 (Draft)*
*Không có dòng code, API, schema, hay thay đổi backend nào được tạo ra từ tài liệu này.*
*Tài liệu này phải được CPO, CFO, CTO phê duyệt, và các mục [MỚI] phải qua quy trình tu chính BRB, trước khi bất kỳ phần nào được triển khai.*
