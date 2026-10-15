# Ma Sói — Đặc tả luật chơi (App quản trò offline, chỉ dùng điện thoại)

Nguồn chính thức: tài liệu "Ma Sói — Đặc tả luật chơi" trên Claude (bản 14/10/2026). File này là bản copy cho agent đọc. Khi luật thay đổi, cập nhật cả hai nơi.

## 1. Tổng quan

App đóng vai **quản trò** cho ván Ma Sói chơi trực tiếp ngoài đời: không cần người quản trò, ai cũng được chơi. **Không có thiết bị quản trò riêng**: mỗi người chỉ dùng điện thoại của mình.

- **Điện thoại người chơi**: mỗi người vào phòng bằng mã, nhận vai bí mật, thực hiện hành động ban đêm và bỏ phiếu ban ngày. Mọi điện thoại đều hiện phần công khai (lượt gọi đêm, đếm giờ, lời công bố, phiếu ban ngày) bằng chữ.
- **Máy làm loa**: điện thoại của **chủ phòng** phát giọng đọc lời dẫn và công bố kết quả; các máy khác im lặng, chỉ hiện chữ. Máy làm loa mất kết nối thì người kế tiếp theo thứ tự ghế (còn kết nối) tự động làm loa; trước khi xếp ghế thì theo thứ tự vào phòng. Phòng ghi nhớ ai đang làm loa.
- **Ban ngày** mọi người nói chuyện trực tiếp; app chỉ lo đếm giờ, đếm phiếu và công bố.

Game có hai phe chính, cộng một phe thứ ba có thể xuất hiện:

| Phe | Thế mạnh | Mục tiêu |
| --- | --- | --- |
| Sói (thiểu số) | Biết rõ đồng bọn | Số sói ≥ số dân còn sống |
| Dân (đa số) | Số đông và các vai đặc biệt | Tiêu diệt hết sói |
| Cặp đôi khác phe (nếu có) | Biết vai của nhau | Chỉ còn hai người họ sống |

Nguyên tắc chung: **không lật vai khi chết** (ngoại lệ duy nhất là Thợ săn lộ vai khi bắn).

## 2. Chuẩn bị ván

Một ván có **5–30 người**. Chủ phòng thiết lập theo thứ tự. **Chủ phòng là người tạo phòng, và cũng là người chơi.** Quyền chủ phòng **chỉ có đến hết bước chia vai**; từ lúc chia vai, chủ phòng là người chơi bình thường, không còn nút điều khiển nào (nhưng điện thoại vẫn làm loa như đã nói ở §1). Chủ phòng rời phòng trước khi chia vai thì quyền chuyển cho người vào phòng kế tiếp (còn trong phòng). Các bước:

1. **Tạo phòng**, người chơi vào bằng mã phòng và nhập nickname.
2. **Xếp chỗ**: sắp tên người chơi theo vòng tròn đúng vị trí ngồi ngoài bàn (cần cho Thám tử và Hiệp sĩ kiếm gỉ).
3. **Chọn vai** theo số người (bảng dưới), chọn phương án Sói nguyền hoặc sói thường. **Danh sách vai của ván là công khai**: mọi người xem được trước khi chia vai (chỉ biết có vai nào, không biết ai giữ).
4. **Chỉnh thời gian** các pha nếu muốn.
5. **Chia vai** ngẫu nhiên; mỗi người xem vai trên điện thoại của mình trong pha **xem vai** có đếm ngược. Hết giờ, ván tự sang bước tiếp (bầu Trưởng làng, rồi đêm 1).

### Giới hạn số sói

Tổng số sói, **tính cả người sẽ bị nguyền**, không vượt 30% số người chơi, làm tròn xuống. Sói tiên tri cũng tính vào giới hạn này.

| Số người | Tổng sói tối đa | Phương án có Sói nguyền | Phương án sói thường |
| --- | --- | --- | --- |
| 5–6 | 1 | Không áp dụng | 1 sói |
| 7–9 | 2 | 1 Sói nguyền | 2 sói |
| 10–13 | 3 | 1 sói + 1 Sói nguyền | 3 sói |
| 14–16 | 4 | 2 sói + 1 Sói nguyền | 4 sói |
| 20 | 6 | 4 sói + 1 Sói nguyền | 6 sói |
| 30 | 9 | 7 sói + 1 Sói nguyền | 9 sói |

### Vai theo số người

| Vai | Điều kiện |
| --- | --- |
| Thám tử | Ván trên 8 người |
| Cupid | Ván từ 10 người |
| Các vai còn lại | Chủ phòng tự chọn, tuỳ số người |

### Số lượng mỗi vai

| Vai | Số lượng một ván |
| --- | --- |
| Dân làng | 1 trở lên |
| Sói (sói thường) | 1 trở lên, trong giới hạn số sói ở trên |
| Mọi vai còn lại, gồm Sói nguyền và Sói tiên tri | tối đa 1 |

Các vai chức năng là **duy nhất**: không có ván nào hai Tiên tri hay hai Bảo vệ.
Giới hạn này có thể mở rộng sau (ví dụ cho ván rất nhiều người), nhưng bản đầu
chốt là tối đa 1.

### Thời gian mặc định (chủ phòng chỉnh được)

| Pha | Thời gian |
| --- | --- |
| Xem vai | 30 giây |
| Bầu Trưởng làng | 1 phút |
| Mỗi vai ban đêm | 30 giây |
| Thông báo sáng / hoàng hôn | 10 giây |
| Ban ngày (thảo luận và đề cử) | 7 phút, có thể vote bỏ qua |
| Đếm ngược chốt đề cử | 10 giây |
| Bào chữa | 1 phút |
| Vote giết hay tha | 30 giây |
| Thợ săn chọn người | 15 giây |
| Trưởng làng chọn người kế nhiệm | 15 giây |

Chủ phòng chỉnh giờ trước khi chia vai. Mỗi pha có giới hạn: từ 5 giây đến 10 phút; ban ngày từ 1 đến 30 phút; thông báo sáng/hoàng hôn và đếm ngược chốt đề cử từ 5 giây đến 1 phút. Giá trị ngoài giới hạn bị từ chối.

### Vào lại phòng

Người chơi hết pin hoặc thoát app giữa ván vào lại phòng, **giữ nguyên vai** và trạng thái. Trong lúc vắng mặt, mọi lượt của họ coi như không thực hiện.

Lúc vào phòng, mỗi người được cấp:

- một **mã bí mật**, điện thoại tự lưu, người chơi không cần biết;
- một **PIN 4 số**, hiện trên màn hình của riêng người đó, kèm lời nhắc ghi nhớ.

Có hai cách vào lại:

1. **Cùng máy**: app dùng mã bí mật đã lưu, vào lại ngay.
2. **Máy khác**: nhập mã phòng, nickname và PIN. Vào lại bằng PIN thành công thì người đó được cấp **mã bí mật mới** cho máy mới, mã cũ hết hiệu lực (máy cũ không vào lại được bằng mã cũ).

Nhập sai PIN **5 lần** thì việc vào lại bằng PIN cho nickname đó bị **khoá 1 phút**. Cách vào lại cùng máy bằng mã bí mật không bị khoá. Chỉ gõ nickname mà không có mã bí mật hay PIN thì không vào lại được.

## 3. Ban đêm

Tất cả nhắm mắt. Loa gọi từng vai bằng giọng nói (mọi điện thoại cũng hiện chữ); vai được gọi mở mắt, thao tác trên điện thoại của mình rồi nhắm lại. App gom mọi lựa chọn và **chỉ tính kết quả lúc trời sáng**.

| Thứ tự | Vai | Khi nào gọi |
| --- | --- | --- |
| 0 | Cupid, rồi cặp đôi | Chỉ đêm đầu |
| 1 | Nguyệt Nữ | Mỗi đêm |
| 2 | Bảo vệ | Mỗi đêm |
| 3 | Sói nguyền | Chỉ 3 đêm đầu |
| 4 | Sói tiên tri | Mỗi đêm |
| 5 | Bầy sói | Mỗi đêm, kể cả đêm đầu |
| 6 | Phù thủy | Mỗi đêm |
| 7 | Tiên tri | Mỗi đêm |
| 8 | Thám tử | Mỗi đêm cho đến khi đã dùng |

- **Lượt Bầy sói luôn có** ở mọi ván và mọi đêm, vì mọi ván đều có sói. Sói thường, Sói nguyền và Sói tiên tri đều cắn được trong lượt này; bầy gồm mọi người đang thuộc phe sói. Không có sói thường trong danh sách vai thì lượt này vẫn được gọi.
- **Chỉ gọi các vai có trong ván** (theo danh sách vai đã chọn, vốn công khai). Vai không có trong ván không bao giờ được gọi.
- **Gọi giả**: vai có trong ván nhưng người giữ đã chết vẫn được gọi với thời gian như thường, để không ai suy ra vai nào còn hay mất.
- **Mỗi lượt gọi chạy đủ thời gian**, thật hay giả. Lượt chỉ kết thúc khi hết giờ, không kết thúc sớm khi vai đã chọn xong (nếu không, lượt thật ngắn hơn lượt giả sẽ lộ vai còn sống). Mọi điện thoại (kể cả máy làm loa) chỉ biết vai được gọi và thời lượng, không biết lượt đó thật hay giả.
- **Hết giờ mà chưa chọn**: coi như không thực hiện. Ví dụ Bảo vệ không chọn thì đêm đó không che ai.
- Bầy sói không cần ngẩng lên nhìn nhau: điện thoại hiện danh sách đồng bọn và phiếu của từng con theo thời gian thực.
- Người bị Nguyệt Nữ ghé thấy thông báo trên màn hình khi đến lượt: “Được Nguyệt Nữ ghé chơi, không thể sử dụng chức năng”.

## 4. Ban ngày

Mỗi ngày treo **tối đa 1 người**. Mọi phiếu ban ngày đều **công khai**: mọi điện thoại hiện rõ ai vote ai.

Người đã chết không được nói, không được vote, và **được xem toàn bộ vai và diễn biến** từ lúc chết. Khi chơi online sau này, người chết bị cấm chat và video call.

Khi **ván kết thúc**, mọi người (cả người còn sống) được xem toàn bộ vai và diễn biến.

### Thông báo buổi sáng

| Tình huống | Loa công bố |
| --- | --- |
| Có người chết | Tên người chết, không nêu vai |
| Không ai chết vì Bảo vệ che đúng người bị cắn | “Không có ai chết, được bảo vệ” |
| Không ai chết vì lý do khác (phù thủy cứu, sói không thống nhất, sói bị chặn) | “Không có ai chết” |

Khi **nhiều người chết cùng một đêm**, loa đọc tên theo **thứ tự chỗ ngồi** (từ ghế đầu tiên), không bao giờ theo thứ tự phân giải, và **không nêu nguyên nhân chết hay vai**. Nhờ vậy người chơi không đoán được ai bị cắn, ai bị độc.

Người chết ban đêm **không được trăng trối**. Nếu người chết là Thợ săn, Thợ săn bắn ngay sau thông báo (mục 6).

### Thảo luận và xử trảm

1. **Thảo luận và đề cử** dùng chung đồng hồ ban ngày **7 phút**. Đồng hồ chỉ dừng trong lúc bào chữa và vote giết hay tha.
    - **Nút bỏ qua**: mỗi người bấm một phiếu (Trưởng làng **không** x2), bật tắt được. Nếu số người bấm **lớn hơn 66%** số người còn sống thì phần thời gian còn lại bị bỏ, chuyển sang đêm. Chỉ khi đề cử **một người** thì bỏ qua của người đó mới tự huỷ; chọn “không ai” thì giữ nguyên.
2. **Đề cử**: mỗi người còn sống chọn một người khác để đưa lên giàn (không được tự đề cử mình), hoặc chọn “không ai”. Mọi người được đổi lựa chọn cho đến khi đề cử đóng.
    - Khi **mọi người còn sống** đã chọn một người hoặc “không ai” và có ít nhất một người được đề cử, bắt đầu **đếm ngược chốt 10 giây**. Trong lúc đếm, ai cũng đổi được lựa chọn, và **mỗi lần có người đổi thì đếm lại từ đầu**. Hết 10 giây không ai đổi thì chốt: người nhiều phiếu nhất lên giàn, kể cả khi số người không đề cử nhiều hơn. Đồng hồ ban ngày **không dừng** trong lúc đếm ngược (nếu không, đổi phiếu liên tục sẽ kéo dài ban ngày mãi); nếu nó hết giữa chừng thì người dẫn đầu lên giàn ngay theo luật hết giờ (bước 6). Nút bỏ qua vẫn bấm được.
    - Nếu mọi người đều chọn “không ai” (hoặc ai đổi khiến không còn ai được đề cử) thì không đếm ngược, thảo luận tiếp. Về sau có người đổi sang đề cử một người thì đếm ngược chốt bắt đầu.
    - Phiếu của Trưởng làng tính **x2**.
    - Hòa đề cử: người có phiếu **còn hiệu lực đến sớm nhất** lên giàn. Phiếu đã đổi hoặc rút không tính; đổi phiếu thì phiếu mới tính từ lúc đổi.
    - Người đã được **tha** trong ngày không thể bị đề cử lại cho đến khi sang đêm.
3. **Bào chữa** 1 phút. Người trên giàn mất quyền vote.
4. **Vote giết hay tha**. Người trên giàn không được vote; ai không vote coi như bỏ phiếu trắng. Phiếu của Trưởng làng tính **x2** ở cả giết lẫn tha. Chỉ treo khi số phiếu giết **lớn hơn** số phiếu tha; hòa thì tha. Vote đóng khi mọi người được vote đã vote hoặc hết 30 giây.
5. Nếu **tha**: phiếu đề cử và phiếu bỏ qua bị xoá sạch, đề cử lại từ đầu (bước 2) nếu đồng hồ ban ngày còn giờ. Danh sách người đã được tha **giữ suốt ngày**. Nếu **treo**: người đó chết, sang đêm. Người bị treo **không lật vai** (§1).
6. **Hết đồng hồ ban ngày**:
    - Nếu có người được đề cử thì người nhiều phiếu nhất vẫn bào chữa 1 phút và bị vote giết/tha; xong sang đêm dù kết quả thế nào.
    - Nếu không ai được đề cử thì sang đêm ngay.

## 5. Vai phe Sói

Mọi vai phe sói (sói thường, Sói nguyền, Sói tiên tri) đều **dậy cắn cùng bầy**, biết mặt đồng bọn, và bị Tiên tri soi ra “sói”.

### Sói

- Mỗi đêm, kể cả đêm đầu, bầy vote chọn 1 người để cắn. App hiện phiếu từng con theo thời gian thực, được đổi phiếu thoải mái.
- Bầy được cắn **bất kỳ người sống nào**, kể cả đồng đội sói và người vừa bị nguyền (cắn đồng đội để lấy lòng tin là nước đi hợp lệ). Thông báo buổi sáng không nêu vai.
- Hết giờ, người có **quá nửa số phiếu** của bầy bị cắn. Không ai quá nửa thì đêm đó không ai bị cắn.
- Con sói bị Nguyệt Nữ ghé không được vote; số phiếu của bầy tính trên những con còn được vote. Nếu đó là con sói duy nhất thì đêm đó không ai bị cắn.

### Sói nguyền

- Được gọi riêng **trước bầy sói**, chỉ trong **3 đêm đầu**. Nguyền **1 lần duy nhất** trong game.
- Việc nguyền có hiệu lực **ngay lập tức**: người bị nguyền nhận thông báo trên điện thoại, và ngay đêm đó họ được gọi dậy cắn cùng bầy. Bầy sói thấy đồng bọn mới trên điện thoại. Khi chơi trực tiếp, mỗi đêm đến lượt Sói nguyền, **tất cả điện thoại đều rung nhẹ**, để không ai biết Sói nguyền có nguyền hay không. Khi chơi online thì chỉ cần thông báo.
- Người bị nguyền chuyển sang phe sói và **giữ nguyên năng lực** của vai cũ (Tiên tri, Bảo vệ, Thợ săn…).
- Bầy vẫn cắn được người vừa bị nguyền như mọi người sống khác (xem “Sói”).
- Bảo vệ **không chặn được** việc nguyền. Tiên tri và Thám tử soi người bị nguyền ra sói.
- Bị Nguyệt Nữ ghé ở đêm 1 hoặc 2: lượt nguyền chưa mất, đêm sau dùng được. Bị ghé ở **đêm 3**: mất nguyền luôn (không gọi sang đêm 4, vì như vậy sẽ lộ là bị ghé).
- Hết 3 đêm chưa nguyền thì mất chức năng, trở thành sói thường.
- App không cho nguyền một con sói.

### Sói tiên tri

- Được gọi riêng **sau Sói nguyền và ngay trước bầy sói**, mỗi đêm soi 1 người. Thứ tự khi có cả ba vai: Sói nguyền → Sói tiên tri → bầy sói.
- Kết quả chỉ có “**là Tiên tri**” hoặc “**không phải Tiên tri**”, và được chia sẻ cho cả bầy (bầy được gọi ngay sau đó).
- Bị Nguyệt Nữ ghé: mất cả lượt soi và lượt vote cắn đêm đó.

## 6. Vai phe Dân

### Dân làng

Không có năng lực đặc biệt, chỉ có lá phiếu ban ngày.

### Tiên tri

Mỗi đêm soi 1 người, biết người đó **là sói** hay **không phải sói**. Mọi vai phe sói và người đã bị nguyền đều ra sói.

### Bảo vệ

- Mỗi đêm che 1 người khỏi bị sói cắn. Được che chính mình.
- Không được che cùng một người hai đêm liên tiếp.
- Không chặn được thuốc độc của Phù thủy và việc nguyền.

### Phù thủy

- Có **1 bình cứu** và **1 bình độc** cho cả game. Được dùng cả hai trong cùng một đêm.
- **Bình cứu**: khi còn bình cứu, phù thủy được biết người bị sói cắn đêm đó và chọn cứu hay không. Được tự cứu mình. Phù thủy **không biết** người đó có được Bảo vệ che hay không, nên có thể phí bình.
- Khi đã hết bình cứu, phù thủy không còn được biết ai bị cắn.
- **Bình độc**: giết 1 người sống bất kỳ, kể cả chính mình. Bảo vệ không chặn được.

### Thợ săn

- Không dậy ban đêm. Khi chết **vì bất kỳ lý do gì** (bị cắn, bị treo, thuốc độc, chết theo người yêu…), được kéo theo 1 người.
- Lúc đó cả bàn im lặng, Thợ săn tự quyết định trong 15 giây. Thợ săn lộ vai khi bắn.
- Nguyệt Nữ không chặn được chức năng này.

### Nguyệt Nữ

- Được gọi đầu tiên mỗi đêm, chọn “ngủ” với 1 người. Không được chọn cùng một người hai đêm liên tiếp.
- Người được ghé **mất chức năng ban đêm** đêm đó và được thông báo trên màn hình. Chỉ ảnh hưởng người đó, không ảnh hưởng cả bầy.
- Không chặn được chức năng ban ngày (Thợ săn).
- Bị sói cắn thì chết bình thường.

### Thám tử

- Chỉ có trong ván **trên 8 người**. Dùng **1 lần** trong game.
- Chọn 1 người; app cho biết trong vùng gồm người đó và 2 người ngồi hai bên có sói hay không (người bị nguyền tính là sói).
- Bị Nguyệt Nữ ghé thì không dùng được đêm đó, nhưng **không mất lượt**.

### Hiệp sĩ kiếm gỉ

- Bị sói cắn thì **chết ngay**, nhưng **con sói còn sống gần nhất về phía bên trái** hiệp sĩ (theo thứ tự chỗ ngồi) trúng độc và chết vào đêm hôm sau.
- Bảo vệ che hiệp sĩ: không ai chết, sói không trúng độc.
- Phù thủy cứu hiệp sĩ: hiệp sĩ sống, nhưng sói **vẫn trúng độc**.

## 7. Cupid và phe thứ ba

- Chỉ có trong ván **từ 10 người**. Cupid thuộc phe Dân nếu không tự ghép mình.
- Đêm đầu, Cupid chọn 2 người bất kỳ (được chọn cả chính mình) thành cặp đôi. Sau đó hai người yêu được gọi dậy để biết mặt và vai của nhau.
- Một người trong cặp chết thì người kia **chết theo**, vì bất kỳ lý do gì.
- Phe của cặp đôi được **tính lại mỗi khi có người trong cặp đổi phe** (bị nguyền).
- Điện thoại cặp đôi hiện **liên kết** giữa hai người, bắt đầu từ lúc Cupid được gọi, đổi màu theo phe hiện tại của cặp (bảng dưới). Cupid không thấy liên kết và không biết phe của cặp.

| Cặp đôi | Thuộc phe | Màu liên kết |
| --- | --- | --- |
| Cùng là dân | Phe Dân | Trắng |
| Cùng là sói | Phe Sói | Đen |
| Một sói, một dân | **Phe thứ ba**: thắng khi chỉ còn hai người họ sống | Trắng sọc đen |

Trong cặp khác phe, con sói **vẫn dậy cắn cùng bầy** như thường.

## 7b. Trưởng làng

- Trưởng làng là **danh hiệu công khai**, không phải thẻ vai. Ai cũng có thể giữ, kể cả sói. Vai gốc của người đó **không đổi**. Danh hiệu **không tính** vào giới hạn 30% số sói.
- Bầu cử diễn ra sau bước xem vai, **trước đêm 1**. Mỗi người chơi bầu **công khai** cho một người (được tự bầu mình) trong **1 phút**. Nhiều phiếu nhất thắng. Hòa thì server chọn ngẫu nhiên một trong những người hòa; giống chia vai, lệnh nhận `Seed` và tầng room sinh seed bằng `crypto/rand`, không ai bầu ai cũng sẽ random.
- Nếu **không ai bầu**, server chọn ngẫu nhiên **một người trong tất cả người chơi**, giống trường hợp hòa: lệnh nhận `Seed`, tầng room sinh bằng `crypto/rand`, không bao giờ lấy từ client.
- Phiếu của Trưởng làng tính **x2** ở cả đề cử ban ngày lẫn vote giết/tha (cả phiếu giết lẫn phiếu tha). Trưởng làng đang bị xử trảm mất quyền vote như mọi người.
- Khi Trưởng làng chết (vì bất kỳ lý do gì), họ có **15 giây** chọn một người còn sống kế nhiệm, thông báo công khai. Hết giờ thì danh hiệu mất. Nếu Thợ săn chết cùng lúc, Thợ săn bắn trước, rồi Trưởng làng chuyển danh hiệu.

## 8. Điều kiện thắng

App kiểm tra điều kiện thắng **sau mỗi lần có người chết hoặc có người đổi phe** (kể cả chết theo người yêu, bị Thợ săn bắn, và bị nguyền).

| Phe | Thắng khi |
| --- | --- |
| Dân | Không còn con sói nào, kể cả người bị nguyền và con sói đang yêu |
| Sói | Số sói còn sống ≥ số dân còn sống. **Không tính** cặp đôi khác phe vào cả hai bên |
| Cặp đôi khác phe | Chỉ còn hai người họ sống |

Người bị nguyền tính là sói từ lúc bị nguyền.

Làm rõ:

- Nếu **tất cả** người chơi chết cùng một lúc thì ván **hòa**.
- Nếu nhiều điều kiện thắng cùng đúng, thứ tự ưu tiên là: **Cặp đôi khác phe > Dân > Sói**. Ví dụ: chỉ còn một cặp khác phe sống thì Cặp đôi thắng, không phải Sói thắng (0 sói ≥ 0 dân không được tính là sói thắng).
- Sói chỉ thắng khi còn **ít nhất một con sói** sống (không tính sói trong cặp khác phe).
- Kiểm tra điều kiện thắng chạy **sau khi Thợ săn bắn** (bước 9 của mục 9) và sau mỗi lần đổi phe (nguyền). Nếu ván đã kết thúc thì **bỏ qua** việc Trưởng làng chọn người kế nhiệm.

## 9. Thứ tự phân giải lúc trời sáng

App tính kết quả đêm theo đúng thứ tự sau:

1. **Nguyệt Nữ**: xác định người bị chặn; hành động của họ đêm đó bị bỏ.
2. **Nguyền**: đã có hiệu lực ngay lúc Sói nguyền chọn (mục 5); app chỉ ghi nhận để kiểm tra thắng.
3. **Cắn**: nếu Bảo vệ che đúng người bị cắn thì không ai chết (ghi nhận “được bảo vệ”). Nếu không, Phù thủy cứu thì sống; còn lại thì chết.
4. **Hiệp sĩ kiếm gỉ**: nếu hiệp sĩ bị cắn và không được Bảo vệ che, con sói còn sống gần nhất bên trái hiệp sĩ trúng độc (kể cả khi Phù thủy cứu hiệp sĩ).
5. **Thuốc độc** của Phù thủy.
6. **Sói trúng độc từ đêm trước** chết.
7. **Chết theo người yêu**.
8. **Công bố** (mục 4), rồi Thợ săn bắn nếu chết, lặp lại bước 7 cho người bị bắn.
9. **Kiểm tra điều kiện thắng**.

### Tình huống đã chốt

| Tình huống | Kết quả |
| --- | --- |
| Bảo vệ che và Phù thủy cứu cùng người bị cắn | Không ai chết, loa báo “được bảo vệ”, bình cứu vẫn bị dùng |
| Thuốc độc vào người được Bảo vệ che | Vẫn chết |
| Nguyệt Nữ ghé con sói duy nhất còn sống | Không ai bị cắn, loa báo “Không có ai chết” |
| Nguyền trúng Tiên tri, Bảo vệ, Thợ săn | Sang phe sói, giữ nguyên năng lực |
| Bầy sói cắn người vừa bị nguyền | Cắn được, người đó chết như bình thường |
| Người trong cặp đôi bị nguyền | Tính lại phe cặp đôi, đổi màu liên kết |
| Thợ săn chết theo người yêu hoặc vì thuốc độc | Vẫn được bắn |

## 10. Ngoài phạm vi bản đầu

- **Chơi online hoàn toàn**: chat, ghép phòng, xử lý rớt mạng, server kiểm soát toàn bộ thông tin bí mật.
- **Bot chơi thay** người rời ván giữa chừng.
- **Luật lật vai khi chết**: sẽ cải tiến sau.
- **Điểm cân bằng và bộ vai gợi ý** theo số người, cảnh báo khi ván lệch phe.
- **AI qua camera** phát hiện mở mắt ban đêm; cần sự đồng ý rõ ràng của người chơi.
- **Đăng ký và đăng nhập bằng tài khoản Google**: tài khoản lâu dài, vào lại từ bất kỳ máy nào. Bản đầu dùng mã bí mật và PIN (§2 "Vào lại phòng").
