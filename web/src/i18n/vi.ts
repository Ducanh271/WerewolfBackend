// The web game's own Vietnamese texts (CLAUDE.md "Conventions"): whole
// sentences with placeholders, never pieces glued together in code, so another
// language is one more file. Texts the server sends (`say`, `notice`,
// `summary`, the guide, error sentences in src/api/errors.ts) are shown as
// given and never copied here. Placeholders `{name}` are filled by fmt().
export const vi = {
  /** Portrait phone overlay (docs/frontend/giao-dien-3d.md §1). */
  rotatePhone: 'Xoay ngang điện thoại',
  appTitle: 'Ma Sói',

  // Before the room (giao-dien-3d.md §10 "Trước sảnh")
  createRoom: 'Tạo phòng',
  joinRoom: 'Vào phòng',
  back: 'Quay lại',
  modeLabel: 'Chế độ chơi',
  nickname: 'Tên của bạn',
  roomCode: 'Mã phòng',
  alreadyInRoom: 'Tôi đã ở trong phòng này',
  pin: 'Mã PIN',
  rejoin: 'Vào lại',
  connecting: 'Đang kết nối…',
  /** No answer from the server at all (http.ts networkError). */
  network: 'Không kết nối được máy chủ, bạn kiểm tra mạng rồi thử lại nhé.',
  /** welcome.protocolVersion is newer than the app (README "Versioning"). */
  outdated: 'Ứng dụng đã cũ, bạn tải lại trang để cập nhật rồi chơi tiếp nhé.',

  // Room modes (rules §1b)
  modeInPerson: 'Chơi quanh bàn',
  modeOnline: 'Chơi online',
  modeInPersonHint: 'Mọi người ngồi cùng một chỗ, một máy làm loa.',
  modeOnlineHint: 'Mỗi người một nơi, mọi máy cùng phát tiếng.',
  modeIs: 'Chế độ: {mode}',
  modeChangeNote: 'Đổi chế độ thì âm thanh về mặc định của chế độ mới, chỗ đã xếp bị bỏ.',

  // PIN (docs/api/screens/lobby.md "Join screen", rules §2 "Vào lại phòng")
  pinTitle: 'Mã PIN của bạn',
  pinNote: 'Hãy tự nhớ mã này, nó sẽ không hiện lại. Đổi sang máy khác thì nhập tên và mã PIN để vào lại phòng.',
  pinDone: 'Tôi đã nhớ',

  // Lobby (temporary list; the 3D lobby is Bước 3b)
  lobbyTitle: 'Sảnh chờ',
  roomIs: 'Phòng {code}',
  peopleCount: '{n} người',
  hostTag: 'Chủ phòng',
  youTag: 'bạn',
  offlineTag: 'mất kết nối',
  kick: 'Mời ra',
  makeHost: 'Trao chủ phòng',
  rolesInGame: 'Vai trong ván',
  noRolesYet: 'Chủ phòng chưa chọn vai.',
  roleCount: '{name} × {n}',
  seatsButton: 'Xếp chỗ',
  rolesButton: 'Chọn vai',
  timersButton: 'Giờ & âm thanh',
  modeButton: 'Đổi chế độ',
  deal: 'Chia vai',
  waitForHost: 'Chờ chủ phòng chia vai',
  leave: 'Rời phòng',
  reconnecting: 'Mất kết nối, đang vào lại…',

  // Settings panels
  save: 'Lưu',
  close: 'Đóng',
  seatsHint: 'Chạm một người rồi chạm người khác để đổi chỗ. Ghế 1 ở đầu, đi theo chiều kim đồng hồ.',
  seatNumber: 'Ghế {n}',
  seatsStale: 'Có người vừa vào hoặc rời phòng, bạn xếp lại nhé.',
  rolesTotal: 'Đã chọn {chosen} vai cho {n} người.',
  rolesLoading: 'Đang tải danh sách vai…',
  more: 'Thêm',
  fewer: 'Bớt',
  soundTitle: 'Âm thanh',
  soundOneSpeaker: 'Một máy làm loa',
  soundAllPhones: 'Mọi máy cùng phát',
  timersTitle: 'Thời gian',
  timerLabel: {
    roleView: 'Xem vai',
    chiefElection: 'Bầu Trưởng làng',
    nightStep: 'Mỗi vai ban đêm',
    phoneCheck: 'Kiểm tra điện thoại',
    loversMeet: 'Cặp đôi mở mắt nhìn nhau',
    announce: 'Thông báo sáng / hoàng hôn',
    day: 'Ban ngày',
    nominationLock: 'Đếm ngược chốt đề cử',
    defense: 'Bào chữa',
    verdict: 'Vote giết hay tha',
    hunterShot: 'Thợ săn chọn người',
    chiefSuccession: 'Trưởng làng chọn người kế nhiệm',
  },

  // In the game (drawn by later steps)
  gameTitle: 'Ván đang diễn ra',
  gameLater: 'Màn này sẽ có ở các bước sau.',
  timeLeft: 'Còn {time}',
} as const;

export type TextKey = keyof typeof vi;

/** Fills `{name}` placeholders of a sentence from vi. */
export function fmt(template: string, vars: Record<string, string | number>): string {
  return template.replace(/\{(\w+)\}/g, (all, key: string) => (key in vars ? String(vars[key]) : all));
}
