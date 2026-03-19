pub mod error;
pub mod msg;
pub mod state;

use cosmwasm_std::{entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, NonceResponse, QueryMsg};
use crate::state::{LOCAL_STATE, NONCES, SHARED_STATE};

#[entry_point]
pub fn instantiate(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    Ok(Response::new().add_attribute("method", "instantiate"))
}

#[entry_point]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::WriteMixed {
            shared_count,
            local_count,
        } => execute_write_mixed(deps, info, shared_count, local_count),
    }
}

fn execute_write_mixed(
    deps: DepsMut,
    info: MessageInfo,
    shared_count: u64,
    local_count: u64,
) -> Result<Response, ContractError> {
    // increment per-sender nonce
    let nonce = NONCES.may_load(deps.storage, &info.sender)?.unwrap_or(0);
    let new_nonce = nonce + 1;
    NONCES.save(deps.storage, &info.sender, &new_nonce)?;

    // shared writes
    for i in 0..shared_count {
        let key = format!("s_{}_{}", nonce, i);
        SHARED_STATE.save(deps.storage, &key, &new_nonce)?;
    }

    // local writes
    for i in 0..local_count {
        let key = format!("l_{}_{}", nonce, i);
        LOCAL_STATE.save(deps.storage, (&info.sender, &key), &new_nonce)?;
    }

    Ok(Response::new()
        .add_attribute("method", "write_mixed")
        .add_attribute("sender", info.sender.to_string())
        .add_attribute("nonce", new_nonce.to_string())
        .add_attribute("shared_writes", shared_count.to_string())
        .add_attribute("local_writes", local_count.to_string()))
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Nonce { address } => {
            let addr = deps.api.addr_validate(&address)?;
            let nonce = NONCES.may_load(deps.storage, &addr)?.unwrap_or(0);
            to_json_binary(&NonceResponse { nonce })
        }
    }
}
