<?php
/**
 * TC-Gaming Thiên Thành Game
 * PHPSample Ví dụ, Trang thử nghiệm, bạn có thể trực tiếp chạy tệp index.php này
 * @author PHOENIX WU
 * @Date 2017/06/19
 * @Version 1.0.1
 */
header('Content-Type:text/json;charset=utf-8');
include_once('class/tcg_class.php');
$tcg_class = new tcg_class();

/**
 * 2.1. CREATE/REGISTER PLAYER API Giao diện tạo/xác nhận người chơi
 * @param $username Tên tài khoản
 * @param $password Mật khẩu tài khoản
 */
echo "--------> TC-Gaming Sample Code Mã nguồn ví dụ Thiên Thành Game <--------\n";
echo print_r('Thực thi --> 2.1. CREATE/REGISTER PLAYER API Giao diện tạo/xác nhận người chơi');
echo "\n";
$username = 'phoenixPHP';
$password = 'phoenix';
$tcg_class-> create_user($username,$password);
echo "\n";
echo "\n";

/**
 * 2.2. UPDATE PASSWORD API Giao diện cập nhật mật khẩu
 * @param $username Tên tài khoản
 * @param $password Mật khẩu tài khoản
 */
echo print_r('Thực thi --> 2.2. UPDATE PASSWORD API Giao diện cập nhật mật khẩu');
echo "\n";
$newpassword = 'phoenix';
$tcg_class-> update_password($username,$newpassword);
echo "\n";
echo "\n";

	
/**
 * 2.3. GET BALANCE API Giao diện lấy số dư
 * @param $username  		Tên tài khoản
 * @param $product_type   	Mã sản phẩm
 */
echo print_r('Thực thi --> 2.3. GET BALANCE API Giao diện lấy số dư');
echo "\n";
$product_type = 7;
$tcg_class-> get_balance($username,$product_type);
echo "\n";
echo "\n";

/**
 * 2.4. FUND TRANSFER API Giao diện chuyển tiền
 * @param $username			Tên tài khoản
 * @param $product_type   	Mã sản phẩm
 * @param int $fund_type  	Giá trị 1=Chuyển vào  Giá trị 2=Chuyển ra
 * @param $amount			Số tiền
 * @param $reference_no		Mã giao dịch
 */
echo print_r('Thực thi --> 2.4. FUND TRANSFER API Giao diện chuyển tiền');
echo "\n";
$username = 'phoenixPHP';
$product_type=7;
$fund_type='2';
$amount=1;
$reference_no='201606190004';
$tcg_class-> user_transfer($username, $product_type, $fund_type, $amount, $reference_no);
echo "\n";
echo "\n";

/**
 * 2.5. CHECK TRANSACTION STATUS API Giao diện kiểm tra trạng thái giao dịch
 * @param $product_type	Mã sản phẩm
 * @param $reference_no	Mã giao dịch
 */
echo print_r('Thực thi --> 2.5. CHECK TRANSACTION STATUS API Giao diện kiểm tra trạng thái giao dịch');
echo "\n";
$product_type=7;
$reference_no='201606190004';
$tcg_class-> check_transfer($product_type, $reference_no);
echo "\n";
echo "\n";

/**
 * 2.6. LAUNCH GAME API Giao diện khởi động game - Game điện tử
 * @param $username
 * @param $product_type
 * @param $gameMode
 * @param $gameCode
 * @param $platform
 */
echo print_r('Thực thi --> 2.6. LAUNCH GAME API Giao diện khởi động game - Game điện tử');
echo "\n";
$product_type = 7;
$gameMode = 1;
$gameCode = 'G00001';
$platform = 'html5';
$tcg_class-> getLaunchGameRng($username, $product_type, $gameMode, $gameCode, $platform);
echo "\n";
echo "\n";

/**
 * 2.6. LAUNCH GAME API Giao diện khởi động game - Game xổ số
 * @param $username
 * @param $product_type
 * @param $game_mode
 * @param $game_code
 * @param $platform
 * @param $view
 */
echo print_r('Thực thi --> 2.6. LAUNCH GAME API Giao diện khởi động game - Game xổ số');
echo "\n";
$product_type = 2;
$game_mode = '1';
$game_code = 'Lobby';
$platform = 'html5';
$view = 'Lobby';
$tcg_class-> getLaunchGameLottery($username, $product_type, $game_mode, $game_code, $platform, $view);
echo "\n";
echo "\n";


/**
 * 2.7. GAME LIST API Giao diện danh sách game
 * @param $product_type Mã sản phẩm
 * @param $platform 	Nền tảng - flash hoặc html5 hoặc all (tất cả)
 * @param $client_type 	Thiết bị - pc:máy tính, phone:điện thoại, web:trình duyệt web, html5:trình duyệt điện thoại
 * @param $game_type 	Loại game - RNG, LIVE, PVP
 * @param $page 		Trang thứ mấy - Nếu không có giá trị, mặc định page = 1
 * @param $page_size 	Số bản ghi hiển thị mỗi trang
 * @return string
 */
echo print_r('Thực thi --> 2.7. GAME LIST API Giao diện danh sách game');
echo "\n";
$product_type = 7;
$platform = "html5";
$client_type = "web";
$game_type = "RNG";
$page = 1 ;
$page_size = 100;
$tcg_class-> getGameList($product_type, $platform, $client_type, $game_type, $page, $page_size);
echo "\n";
echo "\n";



/**
 * 3.1. GET RNG BET DETAILS Giao diện lấy chi tiết cược game điện tử và live casino
 * @param $end_date 		Ngày kết thúc 2016-01-01 00:00:00
 * @param $count 			Số dòng tối đa
 *
 */
echo print_r('Thực thi --> 3.1. GET RNG BET DETAILS Giao diện lấy chi tiết cược game điện tử và live casino');
echo "\n";
$batch_name="201706262010";
$page=1;
$tcg_class-> get_bet_details($batch_name, $page);
echo "\n";
echo "\n";

/**
 * 3.2. GET RNG BET DETAILS BY MEMBER Giao diện lấy chi tiết cược game điện tử và live casino theo tài khoản
 * @param $username		Tên tài khoản	
 * @param $start_date 	Thời gian bắt đầu	
 * @param $end_date 	Thời gian kết thúc
 */
echo print_r('Thực thi --> 3.2. GET RNG BET DETAILS BY MEMBER Giao diện lấy chi tiết cược game điện tử và live casino theo tài khoản');
echo "\n";
$username = 'phoenixPHP';
$start_date='2017-06-26 20:10:00';
$end_date='2017-06-26 23:59:59';
$page=1;
$tcg_class-> get_bet_details_member($username, $start_date, $end_date, $page);
echo "\n";
echo "\n";






die();
?>
