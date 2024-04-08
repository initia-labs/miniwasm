use cosmwasm_std::{to_json_binary, Binary, Deps, Empty, Env, QueryRequest, StdResult};

use crate::msgs::QueryMsg;
use crate::slinky::{GetAllCurrencyPairsRequest, GetAllCurrencyPairsResponse};
use crate::state::Contract;
use protobuf::Message;

impl<'a> Contract {
    fn get_all_currency_pairs(
        &self,
        deps: Deps,
        _env: Env,
    ) -> StdResult<GetAllCurrencyPairsResponse> {
        let request = GetAllCurrencyPairsRequest {
            special_fields: ::protobuf::SpecialFields::new(),
        };
        let bytes = request.write_to_bytes().unwrap();

        let data = Binary::from(bytes);
        let request = QueryRequest::<Empty>::Stargate {
            path: "/slinky.oracle.v1.Query/GetAllCurrencyPairs".to_string(),
            data,
        };
        let response: GetAllCurrencyPairsResponse = deps.querier.query(&request)?;
        Ok(response)
    }
}

impl<'a> Contract {
    pub fn query(&self, deps: Deps, env: Env, msg: QueryMsg) -> StdResult<Binary> {
        match msg {
            // QueryMsg::GetPrice { pair_id } => to_json_binary(&self.get_price(deps, env, pair_id)?),
            // QueryMsg::GetPriceRaw { pair_id } => {
            //     to_json_binary(&self.get_price_raw(deps, env, pair_id)?)
            // }
            QueryMsg::GetAllCurrencyPairs {} => {
                to_json_binary(&self.get_all_currency_pairs(deps, env)?)
            }
        }
    }
}
