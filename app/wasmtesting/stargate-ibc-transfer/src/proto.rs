pub const MSG_TRANSFER_TYPE_URL: &str = "/ibc.applications.transfer.v1.MsgTransfer";

pub struct MsgTransfer<'a> {
    pub source_port: &'a str,
    pub source_channel: &'a str,
    pub denom: &'a str,
    pub amount: &'a str,
    pub sender: &'a str,
    pub receiver: &'a str,
    pub timeout_timestamp: u64,
    pub memo: Option<&'a str>,
}

pub fn encode_msg_transfer(msg: MsgTransfer) -> Vec<u8> {
    let mut out = Vec::new();

    write_string(&mut out, 1, msg.source_port);
    write_string(&mut out, 2, msg.source_channel);
    write_message(&mut out, 3, encode_coin(msg.denom, msg.amount));
    write_string(&mut out, 4, msg.sender);
    write_string(&mut out, 5, msg.receiver);

    if msg.timeout_timestamp != 0 {
        write_uint64(&mut out, 7, msg.timeout_timestamp);
    }

    if let Some(memo) = msg.memo {
        if !memo.is_empty() {
            write_string(&mut out, 8, memo);
        }
    }

    out
}

fn encode_coin(denom: &str, amount: &str) -> Vec<u8> {
    let mut out = Vec::new();
    write_string(&mut out, 1, denom);
    write_string(&mut out, 2, amount);
    out
}

fn write_string(out: &mut Vec<u8>, field_number: u32, value: &str) {
    write_key(out, field_number, 2);
    write_varint(out, value.len() as u64);
    out.extend_from_slice(value.as_bytes());
}

fn write_message(out: &mut Vec<u8>, field_number: u32, value: Vec<u8>) {
    write_key(out, field_number, 2);
    write_varint(out, value.len() as u64);
    out.extend_from_slice(&value);
}

fn write_uint64(out: &mut Vec<u8>, field_number: u32, value: u64) {
    write_key(out, field_number, 0);
    write_varint(out, value);
}

fn write_key(out: &mut Vec<u8>, field_number: u32, wire_type: u8) {
    write_varint(out, ((field_number as u64) << 3) | wire_type as u64);
}

fn write_varint(out: &mut Vec<u8>, mut value: u64) {
    while value >= 0x80 {
        out.push((value as u8) | 0x80);
        value >>= 7;
    }
    out.push(value as u8);
}

#[cfg(test)]
mod tests {
    use super::{encode_msg_transfer, MsgTransfer};

    #[test]
    fn encodes_ibc_transfer_msg() {
        let encoded = encode_msg_transfer(MsgTransfer {
            source_port: "transfer",
            source_channel: "channel-0",
            denom: "uinit",
            amount: "123",
            sender: "init1sender",
            receiver: "cosmos1receiver",
            timeout_timestamp: 1_700_000_000_000_000_000,
            memo: Some("memo"),
        });

        assert_eq!(
            encoded,
            vec![
                10, 8, 116, 114, 97, 110, 115, 102, 101, 114, 18, 9, 99, 104, 97, 110, 110, 101,
                108, 45, 48, 26, 12, 10, 5, 117, 105, 110, 105, 116, 18, 3, 49, 50, 51, 34, 11,
                105, 110, 105, 116, 49, 115, 101, 110, 100, 101, 114, 42, 15, 99, 111, 115, 109,
                111, 115, 49, 114, 101, 99, 101, 105, 118, 101, 114, 56, 128, 128, 168, 177, 227,
                159, 231, 203, 23, 66, 4, 109, 101, 109, 111,
            ]
        );
    }
}
