#![no_std]
#![no_main]

#[panic_handler]
fn panic(_info: &core::panic::PanicInfo) -> ! {
    loop {}
}

#[no_mangle]
pub extern "C" fn test_round_time() {
    unsafe {
        let result = getRoundTime();
        let result: [u8; 1] = [(result & 0xff) as u8];
        finish(result.as_ref().as_ptr(), 1);
    };
}

#[no_mangle]
pub extern "C" fn test_epoch_start_block_time_stamp() {
    unsafe {
        let result = epochStartBlockTimeStamp();
        let result: [u8; 1] = [(result & 0xff) as u8];
        finish(result.as_ref().as_ptr(), 1);
    };
}

#[no_mangle]
pub extern "C" fn test_epoch_start_block_nonce() {
    unsafe {
        let result = epochStartBlockNonce();
        let result: [u8; 1] = [(result & 0xff) as u8];
        finish(result.as_ref().as_ptr(), 1);
    };
}

#[no_mangle]
pub extern "C" fn test_epoch_start_block_round() {
    unsafe {
        let result = epochStartBlockRound();
        let result: [u8; 1] = [(result & 0xff) as u8];
        finish(result.as_ref().as_ptr(), 1);
    };
}

extern {
    fn finish(data: *const u8, len: i32);
    fn getRoundTime() -> i64;
    fn epochStartBlockTimeStamp() -> i64;
    fn epochStartBlockNonce() -> i64;
    fn epochStartBlockRound() -> i64;
}