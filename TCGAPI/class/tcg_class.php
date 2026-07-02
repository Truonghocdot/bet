<?php
/**
 * <code>lt_common.php</code>
 * TC-Gaming Sample Code Mã nguồn ví dụ Thiên Thành Game
 *  
 * @author PHOENIX WU
 * @Date 2017/06/19
 * @Version 1.0.1
 * @Desc Trước tiên xin cảm ơn mọi người đã tham khảo đoạn code PHP đầu tiên của tôi. Nếu có chỗ nào cần cải thiện, xin hãy chỉ giáo nhiều hơn. Mọi ý kiến xin vui lòng gửi về phoenix.w@tc-gaming.com
 *       Sự chỉ dẫn và góp ý của bạn là động lực để chúng tôi phát triển. Tôi là kỹ sư Java, đây là lần đầu tiên tôi lập trình PHP
 * @Change
 * 2020/10/12 Thêm phương thức mã hóa PHP 7.2
 */
class tcg_class{
	/**
	 * Các tham số chung
	 */
    function __construct(){
        $this->url = 'http://www.url.com/doBusiness.do';	//Kết nối API
        $this->merchant_code = '';							//Mã đại lý
        $this->desKey = '';									//Khóa mã hóa
        $this->signKey = '';								//Chữ ký mã hóa
        $this->currency = ''; 								//Loại tiền tệ
    }
	
    /**
     * 2.1. CREATE/REGISTER PLAYER API Giao diện tạo/xác nhận người chơi
     * @param $username Tên tài khoản
     * @param $password Mật khẩu
     * @return array|SimpleXMLElement
     */
    public function create_user($username, $password ){
        $registerParams = array('username' => $username, 'currency' => $this->currency, 'method' => 'cm', 'password' => $password);
        $result = $this->send_require($registerParams);
        return $result;
    }
	
	/**
	 * 2.2. UPDATE PASSWORD API Giao diện cập nhật mật khẩu
	 * @param $username Tên tài khoản
	 * @param $password Mật khẩu
     * @return array|SimpleXMLElement
     */
	public function update_password($username,$password){
		$registerParams = array('username' => $username, 'currency' => $this->currency, 'method' => 'up', 'password' => $password);
		//print_r($getBalanceParams);
        $result = $this->send_require($registerParams);
        return $result;
	}
	
    /**
     * 2.3. GET BALANCE API Giao diện lấy số dư
     * @param $username  		Tên tài khoản
	 * @param $product_type   	Mã sản phẩm
     * @return array|SimpleXMLElement
     */
    public function get_balance($username,$product_type){
        $getBalanceParams = array('username' => $username, 'method' => 'gb' , 'product_type' => $product_type);
		//print_r($getBalanceParams);
        $result = $this->send_require($getBalanceParams);
        return $result;
    }

    /**
     * 2.4. FUND TRANSFER API Giao diện chuyển tiền
     * @param $username			Tên tài khoản
	 * @param $product_type   	Mã sản phẩm
     * @param int $fund_type  	Giá trị 1=Chuyển vào  Giá trị 2=Chuyển ra
     * @param $amount			Số tiền
     * @param $reference_no		Mã giao dịch
     * @return array|SimpleXMLElement
     */
    public function user_transfer($username, $product_type, $fund_type, $amount, $reference_no){
        $getBalanceParams = array('username' => $username, 'method' => 'ft' , 'product_type' => $product_type,'fund_type' => $fund_type,'amount' => $amount,'reference_no' => $reference_no);
		//print_r($getBalanceParams);
        $result = $this->send_require($getBalanceParams);
        return $result;
    }

    /**
     * 2.5. CHECK TRANSACTION STATUS API Giao diện kiểm tra trạng thái giao dịch
     * @param $product_type	Mã sản phẩm
     * @param $reference_no	Mã giao dịch
     * @return array|SimpleXMLElement
     */
    public function check_transfer($product_type, $reference_no){
        $getBalanceParams = array('method' => 'cs' , 'product_type' => $product_type, 'ref_no' => $reference_no);
		//print_r($getBalanceParams);
        $result = $this->send_require($getBalanceParams);
        return $result;
    }

    /**
     * 2.6. LAUNCH GAME API Giao diện khởi động game - Game điện tử
     * @param $username Tên tài khoản
     * @param $gameMode Giá trị 1=Chính thức Giá trị 0=Thử nghiệm
     * @param $gameCode Mã game
     * @return string
     */
    public function getLaunchGameRng($username, $product_type, $game_mode, $game_code, $platform){
		/** RNG GAME Game điện tử **/
		$getBalanceParams = array('username' => $username, 'method' => 'lg' , 'product_type' => $product_type,'game_mode' => $game_mode,'game_code' => $game_code,'platform'=>$platform);
		//print_r($getBalanceParams);
		$result = $this->send_require($getBalanceParams);
        return $result;
    }
	
	/**
	 * 2.6. LAUNCH GAME API Giao diện khởi động game - Game xổ số
     * @param $username 	Tên tài khoản
	 * @param $product_type Mã xổ số là 2 
     * @param $game_mode 	Giá trị 1=Chính thức Giá trị 0=Thử nghiệm
     * @param $game_code 	Mã game
	 * @param $platform 	Nền tảng flash，html5
	 * @param $view 		Giao diện hiển thị
     * @return string
     */
	public function getLaunchGameLottery($username, $product_type, $game_mode, $game_code, $platform, $view){
		/** Lottery GAME Game xổ số **/
		// Chế độ hiện tại chỉ có thể sử dụng Traditional truyền thống và Traditional_Mobile truyền thống_di động
		$lottery_bet_mode = 'Traditional'; 
		$series = array();
		$series[] = array('game_group_code'=>'SSC','prize_mode_id'=>1,'max_series'=>1956,'min_series'=>1700,'max_bet_series'=>1950,'default_series'=>1800);
		$getBalanceParams = array('username'=>$username, 'method'=>'lg', 'product_type'=> $product_type, 'game_code'=>$game_code, 'game_mode'=>$game_mode, 'platform'=>$platform, 'lottery_bet_mode'=>$lottery_bet_mode, 'view'=>$view, 'series'=>$series);
		$result = $this->send_require($getBalanceParams);
        return $result;
    }

	/**
	 * 2.7. GAME LIST API Giao diện danh sách game
     * @param $product_type Mã sản phẩm
     * @param $platform 	Nền tảng - flash hoặc html5 hoặc all (tất cả)
     * @param $client_type 	Thiết bị - pc:máy tính, phone:điện thoại, web:trình duyệt web, html5:trình duyệt điện thoại
	 * @param $game_type 	Loại game - RNG, LIVE, PVP
	 * @param $page 		Trang thứ mấy - Nếu không có giá trị mặc định page = 1
	 * @param $page_size 	Số bản ghi hiển thị mỗi trang
     * @return string
     */
	public function getGameList($product_type, $platform, $client_type, $game_type, $page, $page_size){
		$getBalanceParams = array('method'=>'tgl', 'product_type'=>$product_type, 'platform'=>$platform, 'client_type'=>$client_type, 'game_type'=>$game_type, 'page'=>$page, 'page_size'=>$page_size);
		$result = $this->send_require($getBalanceParams);
        return $result;
	}
	
	/**
	 * 2.8. Player Game Rank API Giao diện bảng xếp hạng người chơi
	 * @param $product_type 	Mã sản phẩm
	 * @param $game_category 	RNG, LIVE đây là thông số bắt buộc, chỉ được sử dụng khi loại sản phẩm không phải là 1, 2 và 5
	 * @param $game_code 		T2KSSC, SD11X5, P00001
	 * @param $start_date 		Ngày bắt đầu 2016-01-01 00:00:00
	 * @param $end_date 		Ngày kết thúc 2016-01-01 00:00:00
	 * @param $count 			Số dòng tối đa
	 *
	 */
	public function getGameRank($product_type, $game_category, $game_code, $start_date, $end_date, $count){
		$getBalanceParams = array('method'=>'pgr', 'product_type'=>$product_type, 'game_category'=>$game_category, 'game_code'=>$game_code, 'start_date'=>$start_date, 'end_date'=>$end_date, 'count'=>$count);
		$result = $this->send_require($getBalanceParams);
        return $result;
	} 

	/**
	 * 3.1. GET RNG BET DETAILS Giao diện lấy chi tiết cược game điện tử và live casino
	 * @param $batch_name 	Mã lô (batch number)
	 * @param $page 		Trang thứ mấy
	 */
    public function get_bet_details($batch_name, $page){
        $time_str = $stime;
        $getBalanceParams = array('method'=>'bd', 'batch_name'=>$batch_name, 'page'=>$page);
        $result = $this->send_require($getBalanceParams);
        return $result;
    }
	
	/**
	 * 3.2. GET RNG BET DETAILS BY MEMBER Giao diện lấy chi tiết cược game điện tử và live casino theo tài khoản
	 * @param $username		Tên tài khoản	
	 * @param $start_date 	Thời gian bắt đầu	
	 * @param $end_date 	Thời gian kết thúc
	 */
	public function get_bet_details_member($username, $start_date, $end_date, $page){
        $getBalanceParams = array('username'=>$username, 'method'=>'bdm', 'start_date'=>$start_date, 'end_date'=>$end_date, 'page'=>$page);
        $result = $this->send_require($getBalanceParams);
        return $result;
    }
	
	/**
	 * 4.1. GET LOTTO TRANSACTIONS BY MEMBER Giao diện lấy lịch sử giao dịch xổ số theo thời gian thực của thành viên
	 * @param $username		Tên tài khoản	
	 * @param $start_date 	Thời gian bắt đầu	
	 * @param $end_date 	Thời gian kết thúc
	 */
	public function getLottoTxByMember($username, $start_date, $end_date, $page){
        $getBalanceParams = array('username'=>$username, 'method'=>'lmb', 'start_date'=>$start_date, 'end_date'=>$end_date, 'page'=>$page);
        $result = $this->send_require($getBalanceParams);
        return $result;
	}
	
	/**
	 * Về cơ bản, cách viết cũng khá giống nhau. Nếu có thắc mắc gì, vui lòng liên hệ nhóm kỹ thuật TCG... Xin lỗi nhóm biên tập viên tôi lười viết quá rồi!
	 */
	

    public function get_lottoCode(){
        $getBalanceParams = array('method' => 'glgl');
        $result = $this->send_require($getBalanceParams);
        return $result;
    }
	
    /**
     * Gửi yêu cầu chung
     * @param $sendParams
     * @return string
     */
    public function send_require($sendParams){
        $params =  $this->encryptText(json_encode($sendParams),$this->desKey);
        $sign = hash('sha256', $params . $this->signKey);
        $data = array('merchant_code' => $this->merchant_code, 'params' => $params , 'sign' => $sign);
        $options = array(
            'http' => array(
                'header'  => "Content-type: application/x-www-form-urlencoded\r\n",
                'method'  => 'POST',
                'content' => http_build_query($data)
            )
        );
        $context  = stream_context_create($options);
        $result = file_get_contents($this->url, false, $context);
        var_dump($result);
        return $result;
    }
	
    /**
     * Cấu trúc tham số mã hóa
     * @param $plainText
     * @param $key
     * @return string
     */
    function encryptText($plainText, $key) {
		// Các phiên bản PHP dưới 7.1
        //$padded = $this->pkcs5_pad($plainText,mcrypt_get_block_size("des", "ecb"));
        //$encText = mcrypt_encrypt("des",$key, $padded, "ecb");
   
		// Các phiên bản PHP từ 7.1 trở lên
		$padded = $this->pkcs5_pad($plainText, 8);
        $encText = openssl_encrypt($padded, 'des-ecb', $key, OPENSSL_RAW_DATA, '');
		
        return base64_encode($encText);
    }

    /**
     * Cấu trúc tham số giải mã
     * @param $plainText
     * @param $key
     * @return string
     */
    function decryptText($encryptText, $key) {
        $cipherText = base64_decode($encryptText);
        // Các phiên bản PHP dưới 7.1
		//$res = mcrypt_decrypt("des", $key, $cipherText, "ecb");
		// Các phiên bản PHP từ 7.1 trở lên
		$res = openssl_decrypt($cipherText, 'des-ecb', $key, OPENSSL_NO_PADDING);
        $resUnpadded = $this->pkcs5_unpad($res);
        return $resUnpadded;
    }

	/**
     * Padding (Điền đầy)
     * @param $text
     * @param $blocksize
     * @return string
     */
    function pkcs5_pad ($text, $blocksize)
    {
        $pad = $blocksize - (strlen($text) % $blocksize);
        return $text . str_repeat(chr($pad), $pad);
    }
	
    function pkcs5_unpad($text)
    {
        $pad = ord($text{strlen($text)-1});
        if ($pad > strlen($text)) return false;
        if (strspn($text, chr($pad), strlen($text) - $pad) != $pad) return false;
        return substr($text, 0, -1 * $pad);
    }
}