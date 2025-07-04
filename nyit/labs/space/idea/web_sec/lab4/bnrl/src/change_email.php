<?php
session_start();
include 'db.php';

// Ensure user token exists
if (!isset($_COOKIE["user_token"]) || !isset($_POST["new_email"])) {
    exit("❌ Invalid request.");
}

// Extract username from user_token cookie
$cookie_data = explode(":", $_COOKIE["user_token"]);
$cookie_username = $cookie_data[0] ?? null;

if (!$cookie_username) {
    exit("❌ Invalid session.");
}

$email = trim($_POST["new_email"]);
var_dump($email, $cookie_username);
// Perform email update for the username stored in the cookie
$query = "UPDATE users SET email = ? WHERE username = ?";
$stmt = $conn->prepare($query);
$stmt->bind_param("ss", $email, $cookie_username);
$stmt->execute();

$target_username = "CryptoBro419"; // Target user for flag
$flag_email = "crypto@scammer.com"; // Email that triggers the flag
$flag = "FLAG{3Ma1L_PwN3D_1337}"; 
// Check if update was successful
if ($stmt->affected_rows > 0) {
    echo "✅ Email updated successfully!";
    if ($cookie_username === $target_username && $email === $flag_email) {
        echo "\n🎉 FLAG UNLOCKED: " . $flag;
    }
} else {
    echo "❌ Update failed.";
}

$stmt->close();
?>

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Change Email Result</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Poppins:wght@300;400;600&display=swap');
        body {
            font-family: 'Poppins', sans-serif;
            background: linear-gradient(135deg, #232f3e 0%, #37475a 100%);
            color: #f8f8f8;
            margin: 0;
            padding: 0;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
        }
        .header-bar {
            width: 100vw;
            height: 8px;
            background: linear-gradient(90deg, #ffb347 0%, #ff6f3c 100%);
            box-shadow: 0 2px 8px 0 #ffb34744;
            margin-bottom: 2.5rem;
        }
        .result-box {
            background: rgba(255, 255, 255, 0.1);
            padding: 40px 30px 30px 30px;
            border-radius: 20px;
            box-shadow: 0px 15px 40px rgba(0, 0, 0, 0.5);
            text-align: center;
            min-width: 320px;
            max-width: 95vw;
            margin: 0 auto;
            animation: fadeInMain 1s cubic-bezier(.77,0,.18,1);
        }
        @keyframes fadeInMain {
            from { opacity: 0; transform: translateY(40px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .result-box p {
            color: #ffe082;
            font-size: 1.2rem;
            margin: 0;
        }
    </style>
</head>
<body>
    <div class="header-bar"></div>
    <div class="result-box">
        <p><?php // ... existing code ... ?></p>
    </div>
</body>
</html>
