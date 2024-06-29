clc;clear;
% The function of the code：
% Generate channel parameters for TN communication and NTN communication
APandSatelliteNums = 64; % total number of APs and satellites 
                        % (consistent with APGenerate.m and SatelliteGenerate.m).
terminalNums = 141; % number of terminals(consistent with TerminalGenerate.m)
coefficient = zeros(APandSatelliteNums,terminalNums); % The Matrix stores all the channel coefficients

% Odd positions store TN communication channel parameters, 
% while even positions store NTN communication channel parameters.
for i = 1:APandSatelliteNums
    if mod(i,2) == 1
        numbers = exprnd(1, terminalNums, 1);
        coefficient(i,:) = round(numbers, 4);
    else
        x = ShadowedRicianRandGen(0.851,2.91,0.278,terminalNums);
        coefficient(i,:) = round(x', 4);
    end
end