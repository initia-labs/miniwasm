use cosmwasm_schema::cw_serde;
use cw_storage_plus::Item;

#[cw_serde]
pub struct Count {
    pub val: u64,
}

pub const COUNT: Item<Count> = Item::new("count");
