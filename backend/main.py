import pandas as pd
import numpy as np
import statsmodels.tsa.stattools as ts 
from statsmodels.tsa.stattools import adfuller
import statsmodels.api as sm

stock1 = pd.read_csv('stock_price_1.csv')
stock2 = pd.read_csv('stock_price_2.csv')
spread = pd.read_csv('spread_of_stock_prices.csv')



# engle granger test
result = ts.coint(stock1, stock2)
cointegration = result[0] 
p_value = result[1]
crit_value = result[2]

# P Value < 0.05 shows that cointegration exists
print(p_value)
# Approx 0.00000332 which is 


# # adfuller test
# spy_adf = adfuller(stock1)
# print(spy_adf[1])
# nflx_adf = adfuller(stock2)
# print(nflx_adf[1])
# spread_adf = adfuller(spy_adf - nflx_adf)
# print(spread_adf[1])
# ratio_adf = adfuller(spy_adf / nflx_adf)
# print(ratio_adf[1])

# # Perform ADF test on 'spy_adf'
# adf_spy = adfuller(spy_adf)
# p_value_spy = adf_spy[1]
# print("ADF p-value for 'spy_adf':", p_value_spy)

# # Perform ADF test on the spread between 'spy_adf' and 'nflx_adf'
# spread = spy_adf - nflx_adf
# adf_spread = adfuller(spread)
# p_value_spread = adf_spread[1]
# print("ADF p-value for the spread:", p_value_spread)

# # Perform ADF test on the ratio of 'spy_adf' and 'nflx_adf'
# ratio = spy_adf / nflx_adf
# adf_ratio = adfuller(ratio)
# p_value_ratio = adf_ratio[1]
# print("ADF p-value for the ratio:", p_value_ratio)


# Regression 
spector_data = spread

spector_data.exog = sm.add_constant(spector_data.exog, prepend=False)

# Fit and summarize OLS model
mod = sm.OLS(spector_data.endog, spector_data.exog)

res = mod.fit()

print(res.summary())