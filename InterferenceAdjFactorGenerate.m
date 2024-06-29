clc;clear;
% The function of the code：
% Generate interference factors for TN communication and NTN communication
APandSatelliteNums = 64;% total number of APs and satellites 
                        % (consistent with APGenerate.m and SatelliteGenerate.m).
terminalNums = 141; % number of terminals(consistent with TerminalGenerate.m) 
itfadj = zeros(APandSatelliteNums,terminalNums);% The Matrix stores all the interference factors

% Odd positions store TN communication interference factors, 
% while even positions store NTN communication interference factors.
for i = 1:APandSatelliteNums
    if mod(i,2) == 1
        numbers = rand(1, terminalNums);
        itfadj(i,:) = numbers;
    else
        x = ones(1,terminalNums) * 0.1;
        itfadj(i,:) = x;
    end
end