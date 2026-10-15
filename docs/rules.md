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

Nguyên tắc chung: **không lật vai khi chết**. Ngoại lệ duy nhất là Thợ săn: vai của Thợ săn được công khai ngay khi anh ta chết, vì **bất kỳ nguyên nhân nào** (mục 6).

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
| Sói (sói thường) | 1 trở lên; **0 sói thường trở lên khi có Sói nguyền hoặc Sói tiên tri**; luôn trong giới hạn số sói ở trên |
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
| Kiểm tra điện thoại (đêm 1–3 khi có Sói nguyền; đêm 1 khi có Cupid) | 5 giây |
| Cặp đôi mở mắt nhìn nhau (đêm 1 khi có Cupid) | 10 giây |
| Thông báo sáng / hoàng hôn | 10 giây |
| Ban ngày (thảo luận và đề cử) | 7 phút, có thể vote bỏ qua |
| Đếm ngược chốt đề cử | 10 giây |
| Bào chữa | 1 phút |
| Vote giết hay tha | 30 giây |
| Thợ săn chọn người | 15 giây |
| Trưởng làng chọn người kế nhiệm | 15 giây |

Chủ phòng chỉnh giờ trước khi chia vai. Mỗi pha có giới hạn: từ 5 giây đến 10 phút (kể cả Kiểm tra điện thoại); ban ngày từ 1 đến 30 phút; thông báo sáng/hoàng hôn và đếm ngược chốt đề cử từ 5 giây đến 1 phút. Giá trị ngoài giới hạn bị từ chối.

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
| 0 | Cupid | Chỉ đêm đầu |
| 0b | Kiểm tra điện thoại (tình yêu) | Đêm 1 khi danh sách vai có Cupid |
| 0c | Cặp đôi mở mắt nhìn nhau | Đêm 1 khi danh sách vai có Cupid |
| 1 | Nguyệt Nữ | Mỗi đêm |
| 2 | Bảo vệ | Mỗi đêm |
| 3 | Sói nguyền | Chỉ 3 đêm đầu |
| 3b | Kiểm tra điện thoại (nguyền) | Đêm 1–3 khi danh sách vai có Sói nguyền (xem dưới) |
| 4 | Sói tiên tri | Mỗi đêm |
| 5 | Bầy sói | Mỗi đêm, kể cả đêm đầu |
| 6 | Phù thủy | Mỗi đêm |
| 7 | Tiên tri | Mỗi đêm |
| 8 | Thám tử | Mỗi đêm, kể cả sau khi đã dùng (xem dưới) |

- **Lượt Bầy sói luôn có** ở mọi ván và mọi đêm, vì mọi ván đều có sói. Sói thường, Sói nguyền và Sói tiên tri đều cắn được trong lượt này; bầy gồm mọi người đang thuộc phe sói. Không có sói thường trong danh sách vai thì lượt này vẫn được gọi.
- **Chỉ gọi các vai có trong ván** (theo danh sách vai đã chọn, vốn công khai). Vai không có trong ván không bao giờ được gọi.
- **Gọi giả**: vai có trong ván nhưng người giữ đã chết vẫn được gọi với thời gian như thường, để không ai suy ra vai nào còn hay mất.
- **Mỗi lượt gọi chạy đủ thời gian**, thật hay giả. Lượt chỉ kết thúc khi hết giờ, không kết thúc sớm khi vai đã chọn xong (nếu không, lượt thật ngắn hơn lượt giả sẽ lộ vai còn sống). Mọi điện thoại (kể cả máy làm loa) chỉ biết vai được gọi và thời lượng, không biết lượt đó thật hay giả.
- **Hết giờ mà chưa chọn**: coi như không thực hiện. Ví dụ Bảo vệ không chọn thì đêm đó không che ai.
- **Kiểm tra điện thoại** (5 giây) có thể chạy ở hai chỗ, mỗi chỗ theo lịch cố định khi vai tương ứng có trong danh sách vai (dù vai đó đã dùng chức năng, đã chết hay bị Nguyệt Nữ chặn), nên không lộ gì:
  - **Sau lượt Cupid (0b)**, chỉ **đêm 1**, khi danh sách vai có Cupid.
  - **Sau lượt Sói nguyền (3b)**, **luôn luôn** trong 3 đêm đầu khi danh sách vai có Sói nguyền.
  - Có cả Cupid lẫn Sói nguyền thì đêm 1 có **hai** lần kiểm tra: một sau Cupid, một sau Sói nguyền.

  Loa đọc: “Mọi người mở mắt xem điện thoại, rồi nhắm mắt lại.” Ở mỗi lần kiểm tra, **mọi người còn sống** nhận **một** thông báo trên điện thoại, gộp hai phần: tình trạng nguyền (nếu có Sói nguyền): “Bạn không bị nguyền” hoặc “Bạn đã bị nguyền”; và tình trạng tình yêu (nếu có Cupid): “Bạn không ở trong cặp đôi” hoặc “Bạn yêu X, vai Y” kèm màu liên kết (§7). Mọi thông báo có hình dạng y hệt nhau (cùng loại, cùng trường), chỉ khác giá trị, nên không ai đoán được ai bị nguyền hay ai đang yêu từ cách hiển thị. Người đã bị nguyền từ đêm trước vẫn thấy “Bạn đã bị nguyền” ở các đêm còn lại trong 3 đêm đầu; phần tình yêu được nhắc lại ở các lần kiểm tra sau Sói nguyền, nên cặp đôi thấy màu liên kết mới ngay sau khi một người bị nguyền. Lúc kiểm tra điện thoại mọi người đều mở mắt, nên cặp đôi **không** được nhìn nhau lúc này.
- **Cặp đôi mở mắt nhìn nhau** (10 giây, đêm 1 khi danh sách vai có Cupid) chạy ngay sau lần kiểm tra điện thoại sau Cupid. Loa đọc: “Cặp đôi hãy mở mắt nhìn nhau.” Chỉ hai người yêu mở mắt; mọi người khác nhắm mắt.
- **Thám tử được gọi mỗi đêm, kể cả sau khi đã dùng năng lực**: từ đó là lượt gọi giả, cùng thời lượng như lượt thật. Ngừng gọi sẽ báo cho cả bàn biết ông ta đã dùng.
- Bầy sói không cần ngẩng lên nhìn nhau: điện thoại hiện danh sách đồng bọn và phiếu của từng con theo thời gian thực.
- Người bị chặn chức năng ban đêm (hiện chỉ do Nguyệt Nữ, sau này có thể có vai chặn khác) thấy thông báo chung trên màn hình khi đến lượt: “Bạn không thể sử dụng chức năng đêm nay”. Thông báo **không nêu ai hay vai nào chặn**; người bị chặn không được biết nguồn chặn.

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
- Danh sách vote của bầy có thêm lựa chọn rõ ràng **“Không cắn ai”**, tính như một mục tiêu. Hết giờ, người có **quá nửa số phiếu** của bầy bị cắn. Nếu “Không cắn ai” được quá nửa, hoặc không ai quá nửa, thì đêm đó không ai bị cắn. Sáng hôm sau loa vẫn đọc đúng câu “Không có ai chết” như mọi đêm không ai chết; không có gì cho biết bầy sói chủ động không cắn.
- Con sói bị Nguyệt Nữ ghé không được vote; số phiếu của bầy tính trên những con còn được vote. Nếu đó là con sói duy nhất thì đêm đó không ai bị cắn.

### Sói nguyền

- Được gọi riêng **trước bầy sói**, chỉ trong **3 đêm đầu**. Nguyền **1 lần duy nhất** trong game.
- Việc nguyền có hiệu lực **ngay lập tức** về phe: người bị nguyền thành sói, ngay đêm đó được gọi dậy cắn cùng bầy và được vote cắn. Điện thoại của người bị nguyền và của bầy sói **chỉ hiện thay đổi ở bước Kiểm tra điện thoại** (§3): người bị nguyền thấy “Bạn đã bị nguyền”, bầy thấy đồng bọn mới. Mọi người khác thấy “Bạn không bị nguyền”, nên không ai biết Sói nguyền có nguyền hay không.
- Người bị nguyền chuyển sang phe sói và **giữ nguyên năng lực** của vai cũ (Tiên tri, Bảo vệ, Thợ săn…).
- Bầy vẫn cắn được người vừa bị nguyền như mọi người sống khác (xem “Sói”).
- Bảo vệ **không chặn được** việc nguyền. Tiên tri và Thám tử soi người bị nguyền ra sói.
- Bị Nguyệt Nữ ghé ở đêm 1 hoặc 2 (hay hết giờ mà không chọn): lượt nguyền chưa mất, đêm sau dùng được. Bị ghé ở **đêm 3**: mất nguyền luôn (không gọi sang đêm 4, vì như vậy sẽ lộ là bị ghé).
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

- Không dậy ban đêm. Khi chết **vì bất kỳ lý do gì** (bị cắn, bị treo, thuốc độc, chết theo người yêu, bị Thợ săn khác bắn…), **vai được công khai ngay** (Thợ săn là vai duy nhất lộ khi chết; chỉ lộ vai, không lộ phe) và anh ta **phải bắn** một người còn sống bất kỳ.
- Thứ tự: loa đọc người chết trước (thông báo buổi sáng, hoặc sau khi treo), rồi đọc “X là Thợ săn”, rồi cả bàn im lặng chờ **15 giây** trong lúc Thợ săn chọn. Chỉ điện thoại của Thợ săn có nút bắn. Bắn xong loa đọc “X bắn Y”.
- Không có lựa chọn “không bắn”. Hết 15 giây mà chưa chọn thì **không ai bị bắn**.
- Người bị bắn chết ngay. Cái chết đó kéo theo cùng chuỗi hậu quả (nếu là Thợ săn khác thì lộ vai và bắn tiếp), rồi mới kiểm tra thắng (mục 9). Vì cú bắn nằm trước bước kiểm tra thắng nên nó có thể đổi kết quả ván.
- Thợ săn kiêm Trưởng làng: bắn trước, rồi mới chuyển danh hiệu (mục 7b).
- Thợ săn bị nguyền vẫn bắn như một con sói. Nguyệt Nữ không chặn được chức năng này.

### Nguyệt Nữ

- Được gọi đầu tiên mỗi đêm (đêm 1 thì sau Cupid và cặp đôi), **bắt buộc** chọn “ngủ” với 1 người sống **không phải chính cô**. Không được chọn cùng một người hai đêm liên tiếp.
- Hết giờ mà chưa chọn: server chọn **ngẫu nhiên** một người sống hợp lệ (không phải cô, không trùng người đêm trước). Không còn ai hợp lệ thì đêm đó không ghé ai. Nguyệt Nữ đã chết thì vẫn được gọi giả, không chọn ngẫu nhiên.
- Người được ghé **mất chức năng ban đêm** đêm đó và thấy thông báo chung “Bạn không thể sử dụng chức năng đêm nay” (không nhắc Nguyệt Nữ, xem mục 3). Chỉ ảnh hưởng người đó, không ảnh hưởng cả bầy.
- Không chặn được chức năng ban ngày (Thợ săn).
- Bị sói cắn thì chết bình thường; lần ghé đêm đó vẫn có hiệu lực.

### Thám tử

- Chỉ có trong ván **trên 8 người**. Dùng **1 lần** trong game.
- Chọn 1 người còn sống (được chọn chính mình). Vùng gồm người đó và **người còn sống gần nhất ở mỗi bên** theo thứ tự chỗ ngồi (người chết bị bỏ qua). App chỉ cho biết “có sói” hoặc “không có sói” trong vùng, **không** cho biết số lượng hay tên.
- Tính theo phe hiện tại: người bị nguyền và sói trong cặp đôi đều tính là sói.
- Bàn còn dưới 3 người sống thì vùng là tất cả người còn sống (không tính trùng).
- Người **đã chết theo luật nhưng chưa được công bố** (sói trúng độc Hiệp sĩ, người đau buồn vì người yêu, xem §7 và §9 bước 6) vẫn tính là **còn sống** với Thám tử: cả khi tìm người ngồi cạnh lẫn khi đếm sói. Kết quả phải khớp với bàn mà Thám tử đang thấy; bỏ qua người đó sẽ lộ là có người sắp chết.
- Kết quả chỉ Thám tử thấy.
- Được gọi mỗi đêm, kể cả sau khi đã dùng (lượt gọi giả, xem §3).
- Bị Nguyệt Nữ ghé thì không dùng được đêm đó, nhưng **không mất lượt**.

### Hiệp sĩ kiếm gỉ

- Năng lực bị động: Nguyệt Nữ không chặn được. Hiệp sĩ bị nguyền vẫn giữ năng lực (bị chính bầy mình cắn vẫn kích hoạt).
- Bị bầy sói cắn (và Bảo vệ không che) thì **chết ngay**, và **con sói còn sống gần nhất về phía bên trái** hiệp sĩ trúng độc. "Bên trái" là ghế kế tiếp theo chiều kim đồng hồ (ghế đánh số tăng dần). Tính cả mọi người phe sói (sói thường, Sói nguyền, Sói tiên tri, sói trong cặp đôi), bỏ qua người chết (kể cả con sói đã chết theo luật nhưng chưa được công bố; việc này chỉ tính bên trong, không ai thấy) và người không phải sói, không bao giờ là chính hiệp sĩ.
- Sói trúng độc sống đúng 24 giờ: đêm hôm sau, ngay khi bước bầy sói bắt đầu, nó **chết ngay**. Điện thoại nó báo cho nó biết và nó không được vote. Cả bầy thấy nó không còn vote được. Nếu nó là sói duy nhất thì đêm đó không ai bị cắn.
- Cái chết đó được loa công bố **sáng hôm sau cùng các cái chết khác** (theo thứ tự chỗ ngồi, không nêu nguyên nhân hay vai). Người ngoài bầy không biết gì cho đến lúc đó, nên về mặt hiển thị nó vẫn "còn sống" cho đến khi loa đọc. Nếu nó là Thợ săn (bị nguyền) hay Trưởng làng thì Thợ săn bắn và Trưởng làng chọn người kế nhiệm sau thông báo, như mọi cái chết khác.
- Bảo vệ che hiệp sĩ: không ai chết, sói không trúng độc.
- Phù thủy cứu hiệp sĩ: hiệp sĩ sống, nhưng sói **vẫn trúng độc**.
- Chỉ bị **bầy sói cắn** mới kích hoạt. Bị Phù thủy đầu độc, bị treo, bị Thợ săn bắn thì không có gì xảy ra.
- Sói trúng độc chết sớm hơn (bị treo, bị Phù thủy đầu độc, bị Thợ săn bắn) thì không có gì xảy ra thêm.
- Chỉ xét thắng thua vào ban ngày (mục 8).

## 7. Cupid và phe thứ ba

- Chỉ có trong ván **từ 10 người**. Cupid thuộc phe Dân nếu không tự ghép mình.
- Chỉ gọi **đêm đầu** (thứ tự 0, đầu tiên trong đêm). Cupid chọn 2 người sống **khác nhau** (được chọn cả chính mình) thành cặp đôi. Hết giờ mà chưa chọn: server ghép **2 người sống ngẫu nhiên** (seed do tầng room sinh bằng `crypto/rand`).
- Hai người yêu biết người kia và vai của họ ở bước **Kiểm tra điện thoại** ngay sau lượt Cupid ở đêm 1 (§3), không phải lúc Cupid chọn. Ngay sau đó là bước **Cặp đôi mở mắt nhìn nhau** (10 giây): chỉ hai người yêu mở mắt nhìn nhau. Bầy sói **không** biết một con của mình đang yêu.
- **Phe cá nhân không đổi**: sói trong cặp khác phe vẫn là Phe Sói, vẫn cắn cùng bầy. Phe của cặp đôi được tính ở cấp cặp và **tính lại mỗi khi có người trong cặp đổi phe** (bị nguyền): hai người yêu thấy màu mới ở lần Kiểm tra điện thoại kế tiếp.
- Điện thoại cặp đôi hiện **liên kết** giữa hai người (người kia, vai của họ, màu theo phe hiện tại của cặp, bảng dưới). Cupid không thấy liên kết và không biết phe của cặp.
- Một người trong cặp chết thì người kia **chết theo**, vì bất kỳ lý do gì, nhưng thời điểm phụ thuộc lúc người đầu tiên chết:
  - **Chết ban đêm** (cắn, thuốc độc, độc Hiệp sĩ): người kia chết **cùng đêm đó**, ở bước 7 của §9.
  - **Chết ban ngày** (bị treo, bị Thợ săn bắn, kể cả Thợ săn bắn lúc trời sáng): người kia **chưa chết ngay** mà “đau buồn mà chết” trong **đêm kế tiếp**. Đến lúc đó họ vẫn là người sống: được vote, đề cử, có thể bị treo (bị treo thì chết luôn), kiểm tra thắng thua tính họ còn sống, và trên máy người khác họ hiện còn sống đến khi loa đọc tên.
  - Vai **có lượt gọi ban đêm** chết ngay khi đến lượt **đầu tiên** của mình trong đêm đó (Sói nguyền chết ở lượt Sói nguyền, không đợi tới lượt bầy; người phe sói còn có lượt bầy thì lượt nào đến trước). Điện thoại báo cho họ, họ không dùng được chức năng, lượt gọi vẫn chạy đủ giờ. Sói chết kiểu này không được tính vào phiếu bầy.
  - Vai **không có lượt gọi** (Dân, Thợ săn, Hiệp sĩ…) chết ở bước 7 khi tính kết quả đêm.
  - Cái chết được loa công bố **sáng hôm sau cùng người chết khác** (theo thứ tự ghế, không nêu nguyên nhân). Thợ săn chết kiểu này bắn sau thông báo sáng; Trưởng làng truyền chức sau khi xét thắng thua.
  - Nếu cả hai cùng chết trong cùng chuỗi (ví dụ Thợ săn bị treo bắn chính người yêu của mình) thì không còn ai đau buồn.

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

App **chỉ** kiểm tra và công bố thắng thua **vào ban ngày**, ở đúng ba thời điểm:

- (a) **sau thông báo buổi sáng**, kể cả khi đêm đó không ai chết (ví dụ vừa có người bị nguyền làm sói ≥ dân);
- (b) **sau khi treo cổ** một người;
- (c) **sau phát bắn của Thợ săn**.

Luôn xét **sau khi cả chuỗi kết thúc**, không xét giữa chừng: Thợ săn bắn → chết theo người yêu → Thợ săn khác bắn → ... Trưởng làng chỉ truyền chức **sau khi đã xét**, và chỉ khi ván chưa kết thúc. **Không bao giờ kết thúc ván giữa đêm**: dù có người bị nguyền hay chết theo người yêu trong đêm, đêm vẫn chạy đủ mọi bước.

Điều kiện được xét theo tình hình **lúc công bố**, không chốt từ lúc sự việc xảy ra. Ví dụ: nguyền xong mà sói ≥ dân, nhưng Phù thủy đầu độc một con sói cùng đêm, thì sáng ra sói < dân và ván tiếp tục.

| Phe | Thắng khi |
| --- | --- |
| Dân | Không còn con sói nào, kể cả người bị nguyền và con sói đang yêu |
| Sói | Số sói còn sống ≥ số dân còn sống. **Không tính** cặp đôi khác phe vào cả hai bên |
| Cặp đôi khác phe | Chỉ còn hai người họ sống |

Người bị nguyền tính là sói trong mọi lần kiểm tra sau khi bị nguyền.

Làm rõ:

- Nếu **tất cả** người chơi chết cùng một lúc thì ván **hòa**.
- Nếu nhiều điều kiện thắng cùng đúng, thứ tự ưu tiên là: **Cặp đôi khác phe > Dân > Sói**. Ví dụ: chỉ còn một cặp khác phe sống thì Cặp đôi thắng, không phải Sói thắng (0 sói ≥ 0 dân không được tính là sói thắng).
- Sói chỉ thắng khi còn **ít nhất một con sói** sống (không tính sói trong cặp khác phe).
- Sói thắng trong lúc cặp khác phe **vẫn còn sống cả hai**: con sói trong cặp **thua cùng cặp đôi** (màn hình kết thúc ván không tính nó là người thắng).
- Kiểm tra điều kiện thắng chạy **sau khi Thợ săn bắn** (bước 9 của mục 9), luôn vào ban ngày. Nếu ván đã kết thúc thì **bỏ qua** việc Trưởng làng chọn người kế nhiệm.

## 9. Thứ tự phân giải lúc trời sáng

App tính kết quả đêm theo đúng thứ tự sau:

1. **Nguyệt Nữ**: xác định người bị chặn; hành động của họ đêm đó bị bỏ.
2. **Nguyền**: đã có hiệu lực ngay lúc Sói nguyền chọn (mục 5); bước này không làm gì thêm. Thắng thua do nguyền (sói ≥ dân) chỉ xét ở bước 9.
3. **Cắn**: nếu bầy chọn “Không cắn ai” (được quá nửa phiếu) hoặc không mục tiêu nào quá nửa thì không ai bị cắn, không có gì để Bảo vệ che hay Phù thủy cứu, và loa đọc “Không có ai chết” (không phải “được bảo vệ”). Ngược lại, nếu Bảo vệ che đúng người bị cắn thì không ai chết (ghi nhận “được bảo vệ”). Nếu không, Phù thủy cứu thì sống; còn lại thì chết.
4. **Hiệp sĩ kiếm gỉ**: nếu hiệp sĩ bị cắn và không được Bảo vệ che, con sói còn sống gần nhất bên trái hiệp sĩ trúng độc (kể cả khi Phù thủy cứu hiệp sĩ). Con sói này chỉ bị đánh dấu; nó chết ở bước 6 của đêm hôm sau.
5. **Thuốc độc** của Phù thủy.
6. **Chết muộn có lượt gọi**: sói trúng độc từ đêm trước, và người đau buồn vì người yêu chết ban ngày mà có lượt gọi. Về luật họ đã chết từ lúc bước của họ bắt đầu (sói trúng độc: bước bầy sói; người đau buồn: lượt đầu tiên của mình; không vote, không dùng chức năng, không tính vào bầy), nhưng cái chết chỉ được công bố ở bước 8 cùng các cái chết khác.
7. **Chết theo người yêu**: người đau buồn không có lượt gọi chết ở bước này; và người yêu của **bất kỳ ai chết trong đêm** (bước 3 đến 6) chết cùng đêm.
8. **Công bố** (mục 4), rồi Thợ săn chết thì lộ vai và bắn, lặp lại bước 7 cho người bị bắn (và cho Thợ săn bị bắn).
9. **Kiểm tra điều kiện thắng** theo tình hình lúc này (mục 8). Kể cả khi đêm đó không ai chết (ví dụ vừa có người bị nguyền).

### Tình huống đã chốt

| Tình huống | Kết quả |
| --- | --- |
| Bảo vệ che và Phù thủy cứu cùng người bị cắn | Không ai chết, loa báo “được bảo vệ”, bình cứu vẫn bị dùng |
| Thuốc độc vào người được Bảo vệ che | Vẫn chết |
| Bầy chọn “Không cắn ai” được quá nửa phiếu | Không ai bị cắn, loa báo “Không có ai chết”, giống mọi đêm không ai chết; Bảo vệ đã che ai cũng không báo “được bảo vệ” |
| Nguyệt Nữ ghé con sói duy nhất còn sống | Không ai bị cắn, loa báo “Không có ai chết” |
| Nguyền trúng Tiên tri, Bảo vệ, Thợ săn | Sang phe sói, giữ nguyên năng lực |
| Bầy sói cắn người vừa bị nguyền | Cắn được, người đó chết như bình thường |
| Người trong cặp đôi bị nguyền | Tính lại phe cặp đôi, đổi màu liên kết |
| Thợ săn chết theo người yêu hoặc vì thuốc độc | Vẫn được bắn |
| Một người yêu bị treo hoặc bị Thợ săn bắn ban ngày | Người kia sống qua hết ngày, chết đêm kế tiếp (§7), công bố sáng hôm sau |
| Cặp khác phe còn sống cùng Dân và sói khác | Cặp không tính vào hai bên; sói trong cặp vẫn chặn Dân thắng; cặp chỉ thắng khi là hai người cuối |

## 10. Ngoài phạm vi bản đầu

- **Chơi online hoàn toàn**: chat, ghép phòng, xử lý rớt mạng, server kiểm soát toàn bộ thông tin bí mật.
- **Bot chơi thay** người rời ván giữa chừng.
- **Luật lật vai khi chết**: sẽ cải tiến sau.
- **Điểm cân bằng và bộ vai gợi ý** theo số người, cảnh báo khi ván lệch phe.
- **AI qua camera** phát hiện mở mắt ban đêm; cần sự đồng ý rõ ràng của người chơi.
- **Đăng ký và đăng nhập bằng tài khoản Google**: tài khoản lâu dài, vào lại từ bất kỳ máy nào. Bản đầu dùng mã bí mật và PIN (§2 "Vào lại phòng").
