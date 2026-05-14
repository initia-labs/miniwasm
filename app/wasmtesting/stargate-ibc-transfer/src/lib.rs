mod msg;
mod proto;

use cosmwasm_std::{
    entry_point, to_json_binary, Binary, CosmosMsg, Deps, DepsMut, Empty, Env, MessageInfo,
    Response, StdResult,
};

use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::proto::{encode_msg_transfer, MsgTransfer, MSG_TRANSFER_TYPE_URL};

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> StdResult<Response> {
    Ok(Response::new().add_attribute("method", "instantiate"))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    _deps: DepsMut,
    env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> StdResult<Response> {
    match msg {
        ExecuteMsg::IbcTransfer {
            source_channel,
            receiver,
            denom,
            amount,
            timeout_timestamp,
            memo,
        } => {
            let sender = env.contract.address.to_string();
            let amount = amount.to_string();
            let value = encode_msg_transfer(MsgTransfer {
                source_port: "transfer",
                source_channel: &source_channel,
                denom: &denom,
                amount: &amount,
                sender: &sender,
                receiver: &receiver,
                timeout_timestamp,
                memo: memo.as_deref(),
            });

            #[allow(deprecated)]
            let msg = CosmosMsg::<Empty>::Stargate {
                type_url: MSG_TRANSFER_TYPE_URL.to_string(),
                value: Binary::from(value),
            };

            Ok(Response::new()
                .add_message(msg)
                .add_attribute("action", "ibc_transfer")
                .add_attribute("source_port", "transfer")
                .add_attribute("source_channel", source_channel)
                .add_attribute("sender", sender)
                .add_attribute("receiver", receiver)
                .add_attribute("denom", denom)
                .add_attribute("amount", amount))
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(_deps: Deps, _env: Env, _msg: QueryMsg) -> StdResult<Binary> {
    to_json_binary(&())
}
