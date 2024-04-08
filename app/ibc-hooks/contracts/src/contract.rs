use cosmwasm_std::{
    to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdError, StdResult,
};

use crate::msg::{ExecuteMsg, IBCLifecycleComplete, InstantiateMsg, QueryMsg, SudoMsg};
use crate::state::{Count, COUNT};

#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;

// Note, you can use StdResult in some functions where you do not
// make use of the custom errors
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, StdError> {
    let count = Count { val: 0 };

    COUNT.save(deps.storage, &count)?;

    Ok(Response::default())
}

// And declare a custom Error variant for the ones where you will want to make use of it
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, StdError> {
    match msg {
        ExecuteMsg::Increase {} => execute_increase(deps, env, info),
    }
}

pub fn execute_increase(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
) -> Result<Response, StdError> {
    let mut count = COUNT.load(deps.storage)?;
    count.val += 1;
    COUNT.save(deps.storage, &count)?;
    Ok(Response::new())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Get {} => to_json_binary(&query_get(deps)?),
    }
}

fn query_get(deps: Deps) -> StdResult<u64> {
    Ok(COUNT.load(deps.storage)?.val)
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, StdError> {
    match msg {
        SudoMsg::IBCLifecycleComplete(inner_msg) => match inner_msg {
            IBCLifecycleComplete::IBCAck {
                channel: _,
                ack: _,
                sequence,
                success,
            } => {
                let mut count = COUNT.load(deps.storage)?;
                if success {
                    count.val += sequence;
                } else {
                    count.val += 1;
                }

                COUNT.save(deps.storage, &count)?;
                Ok(Response::new())
            }
            IBCLifecycleComplete::IBCTimeout {
                channel: _,
                sequence,
            } => {
                let mut count = COUNT.load(deps.storage)?;
                count.val += sequence;

                COUNT.save(deps.storage, &count)?;
                Ok(Response::new())
            }
        },
    }
}
