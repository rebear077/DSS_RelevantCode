clc; clear;
% The function of the code：
% Generate Points as terminals

% Parameter settings
radiusE = 6371.*1000;
center = [0, 0, radiusE];  % The Earth's radius of 6371 km converts to 6,371,000 meters.
radius = 5000;  % 5 km converts to 5,000 meters.
areaDensity = 2./(10.^(6));  % The point density per square kilometer, divided by 10^6, represents the density per square meter.

points = generatePPPonSmallSphere(center, radius, areaDensity);

disp('Generated points:');
disp(points);

function points = generatePPPonSmallSphere(center, radius, areaDensity)
    % center: Sphere center coordinates
    % radius: Radius of the small sphere
    % areaDensity: Point density per square kilometer
    
    % calculate the surface area of a spherical cap
    area = pi * radius^2;
    
    % Calculate the parameters of the Poisson distribution based on density and area.
    lambda = areaDensity * area;
    
    % Determine the number of points using the Poisson distribution.
    numPoints = poissrnd(lambda);
    
    % Generate the polar angle and azimuth angle of the points.
    theta = 2 * pi * rand(numPoints, 1);  % 方位角，0到2pi
    phi = acos(2 * rand(numPoints, 1) - 1);  % 极角，0到pi
    
    % Coordinates of points on the surface of a small sphere.
    x = center(1) + radius * sin(phi) .* cos(theta);
    y = center(2) + radius * sin(phi) .* sin(theta);
    z = center(3) + radius * cos(phi);
    
    % Merge coordinates into an N*3 matrix.
    points = [x, y, z];
end