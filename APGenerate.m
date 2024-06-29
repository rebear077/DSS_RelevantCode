clc; clear
% The function of the code：
% In a three-dimensional coordinate system, determine a central point. 
% Randomly scatter points within a defined range to obtain the three-dimensional coordinates of the random points.
% Use the obtained random points as ground AP (Access Points).

radiusE = 6371.*1000; % The radius of the Earth.

center = [0, 0, radiusE];  % The radius of the Earth is 6,371 km, which converts to 6,371,000 meters.
smallRadius = 5000;  % The radius of the sphere. (meters).
numPoints = 32;  % The number of points generated

points = generatePointsOnSmallSphere(numPoints, center, smallRadius);

disp('Generated points:');
disp(points);

%% Data storage.
dataFileName = ['./data/AP_', datestr(now,"yyyymmddHHMMSS"), '.mat'];
save(dataFileName,"points");

function points = generatePointsOnSmallSphere(numPoints, center, smallRadius)
    % numPoints: The number of points to be generated.
    % center: The coordinates of the center point, for example [0, 0, radiusE].
    % smallRadius: The radius of the sphere.
    
    % The polar angle (θ) and azimuthal angle (φ) of the generated points.
    theta = 2 * pi * rand(numPoints, 1);  % The azimuthal angle, ranging from 0 to \( 2\pi \) radians (or 0 to 360 degrees).
    phi = acos(2 * rand(numPoints, 1) - 1);  % The polar angle, ranging from 0 to \( \pi \) radians (or 0 to 180 degrees).
    
    % Coordinates of points on the surface of a sphere.
    x = center(1) + smallRadius * sin(phi) .* cos(theta);
    y = center(2) + smallRadius * sin(phi) .* sin(theta);
    z = center(3) + smallRadius * cos(phi);
    
    % Merge coordinates into an N*3 matrix.
    points = [x, y, z];
end
