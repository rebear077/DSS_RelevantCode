clc; clear;
% The function of the code：
% Generate Points as satellites
radiusE = 6371.*1000; 

heightS=1000.*1000;%Satellite altitude of 1000 km.
radiusS= heightS+radiusE; % The radius of the satellite sphere.
densityS =(1.5.*10.^(-7))./(10.^(6)); %The density of the satellite.

min_elevation_BS = 10; % The minimum elevation angle (in degrees).

% Calculate the number of points on the surface of the satellite sphere.
areaS = 4 * pi * radiusS^2;
numPointsS = poissrnd(densityS * areaS);

% Find a point on the Earth's surface to serve as the center point of the ground base station.
pointA =  [0, 0, radiusE];

% Generate points on the surface of the satellite sphere.
[xS, yS, zS] = generateSpherePoints(numPointsS, radiusS);

% Calculate the distance from each point to point A.
distances = sqrt((xS - pointA(1)).^2 + (yS - pointA(2)).^2 + (zS - pointA(3)).^2);
% Calculate the maximum distance from each point to point A.
Distance_max=sqrt((radiusE.*sin(min_elevation_BS).^2+ radiusS^2- radiusE^2)) - radiusE.*sin(min_elevation_BS);

% Find the indices of the points whose distance to point A is less than or equal to d 
% (satellites within the visible range of point A).
visible_indices = find(distances <= Distance_max);

% 提取可视范围内的卫星坐标信息
visible_satellite_x = xS(visible_indices);
visible_satellite_y = yS(visible_indices);
visible_satellite_z = zS(visible_indices);

% Extract the coordinates of the satellites within the visible range.
disp('The coordinates of the satellites within the visible range.:');
for i = 1:numel(visible_indices)
    fprintf('The coordinates of the satellite %d ：(%f, %f, %f)\n', i, visible_satellite_x(i), visible_satellite_y(i), visible_satellite_z(i));
end

% Sphere point generation function
function [x, y, z] = generateSpherePoints(numPoints, radius)
    theta = 2 * pi * rand(numPoints, 1);% Azimuth angle
    phi = acos(2 * rand(numPoints, 1) - 1);% Polar angle
    x = radius * sin(phi) .* cos(theta);
    y = radius * sin(phi) .* sin(theta);
    z = radius * cos(phi);
end
