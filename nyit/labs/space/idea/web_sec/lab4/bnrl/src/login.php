<?php
session_start();
include 'db.php';

if ($_SERVER["REQUEST_METHOD"] == "POST") {
    $username = $_POST["username"];
    $password = $_POST["password"];

    $query = "SELECT id, password, role FROM users WHERE username = '$username'";
    $result = $conn->query($query);

    if ($result->num_rows > 0) {
        $row = $result->fetch_assoc();
        if (password_verify($password, $row["password"])) {
            $_SESSION["user_id"] = $row["id"];
            $_SESSION["role"] = $row["role"];
            $_SESSION["username"] = $username;

             // Generate a random session token
             $random_string = substr(str_shuffle("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"), 0, 6);
             $user_token = $username . ":" . $row["id"] . ":" . $random_string;
             
             $date = date("Ymd"); 
             $role_string = $row["role"] . $date;
             $encoded_role = base64_encode($role_string);
             setcookie("X-Role", $encoded_role, time() + 3600, "/");

             $random_session_track = base64_encode(random_bytes(8));
            setcookie("X-Session-Track", $random_session_track, time() + 3600, "/");

            // **New: Fake User Preferences Cookie**
            $user_prefs = "dark_mode:1;ads:off";
            setcookie("X-User-Pref", base64_encode($user_prefs), time() + 3600, "/");
             // Set cookie (only used for email change verification)
            setcookie("user_token", $user_token, time() + 3600, "/");

            // Redirect based on role
            if ($row["role"] == "admin") {
                header("Location: admin_dashboard.php");
            } else {
                header("Location: user_dashboard.php");
            }
            exit();
        } else {
            $error = "Invalid username or passwordd.";
        }
    } else {
        $error = "Invalid username or password.";
    }
}
?>

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - Vulnerable Web App</title>
    <link href="styles.css" rel="stylesheet">  
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Poppins:wght@300;400;600&display=swap');
        body {
            font-family: 'Poppins', sans-serif;
            background: linear-gradient(135deg, #232f3e 0%, #37475a 100%);
            color: #f8f8f8;
            margin: 0;
            padding: 0;
            min-height: 100vh;
        }
        .header-bar {
            width: 100vw;
            height: 8px;
            background: linear-gradient(90deg, #ffb347 0%, #ff6f3c 100%);
            box-shadow: 0 2px 8px 0 #ffb34744;
            margin-bottom: 2.5rem;
        }
        .container {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(15px);
            padding: 40px 30px 30px 30px;
            border-radius: 20px;
            box-shadow: 0px 15px 40px rgba(0, 0, 0, 0.5);
            text-align: center;
            width: 380px;
            max-width: 95vw;
            margin: 60px auto 0 auto;
            animation: fadeInMain 1s cubic-bezier(.77,0,.18,1);
        }
        @keyframes fadeInMain {
            from { opacity: 0; transform: translateY(40px); }
            to { opacity: 1; transform: translateY(0); }
        }
        h2 {
            color: #ffe082;
            font-size: 2rem;
            margin-bottom: 1.2em;
            letter-spacing: 1px;
        }
        label {
            display: block;
            text-align: left;
            margin-bottom: 6px;
            margin-top: 18px;
            color: #ffe082;
            font-weight: 600;
        }
        input[type="text"], input[type="password"] {
            width: 95%;
            padding: 12px;
            margin: 8px 0 18px 0;
            border: none;
            border-radius: 10px;
            font-size: 16px;
            background: rgba(255, 255, 255, 0.2);
            color: white;
            transition: background 0.3s, transform 0.2s;
        }
        input[type="text"]:hover, input[type="password"]:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: scale(1.02);
        }
        input[type="submit"] {
            background: linear-gradient(90deg, #ffb347 0%, #ff6f3c 100%);
            color: #232f3e;
            padding: 14px 20px;
            border: none;
            border-radius: 12px;
            cursor: pointer;
            width: 100%;
            font-size: 18px;
            font-weight: 600;
            margin-top: 10px;
            transition: background 0.3s, transform 0.2s, box-shadow 0.2s;
            box-shadow: 0 2px 8px 0 #ffb34744;
        }
        input[type="submit"]:hover {
            background: linear-gradient(90deg, #ff6f3c 0%, #ffb347 100%);
            color: #fff;
            transform: scale(1.05);
            box-shadow: 0 4px 16px 0 #ffb34777;
        }
        .forgot-btn {
            display: inline-block;
            background: linear-gradient(90deg, #0077cc, #0055aa);
            color: white;
            padding: 10px 15px;
            border-radius: 8px;
            text-decoration: none;
            font-weight: bold;
            transition: background 0.3s, transform 0.2s;
            margin-top: 18px;
        }
        .forgot-btn:hover {
            background: linear-gradient(90deg, #0055aa, #003388);
            transform: scale(1.05);
        }
        .error {
            color: #ff4f4f;
            font-size: 15px;
            margin-top: 10px;
        }
        .create-btn {
            margin-top: 18px;
            padding: 12px 18px;
            border: none;
            background: linear-gradient(90deg, #ffb347 0%, #ff6f3c 100%);
            color: #232f3e;
            border-radius: 12px;
            cursor: pointer;
            font-size: 18px;
            font-weight: 600;
            width: 100%;
            transition: background 0.3s, transform 0.2s, box-shadow 0.2s;
            box-shadow: 0 2px 8px 0 #ffb34744;
        }
        .create-btn:hover {
            background: linear-gradient(90deg, #ff6f3c 0%, #ffb347 100%);
            color: #fff;
            transform: scale(1.05);
            box-shadow: 0 4px 16px 0 #ffb34777;
        }
    </style>
</head>
<body>
    <div class="header-bar"></div>
    <div class="container">
        <h2>Login</h2>
        <form action="login.php" method="post">
            <label>Username:</label>
            <input type="text" name="username" placeholder="Enter your username" required>
            
            <label>Password:</label>
            <input type="password" name="password" placeholder="Enter your password" required>
            
            <input type="submit" name="login" value="Login">
        </form>
        <p><a href="forgot_password.php" class="forgot-btn">Forgot Password?</a></p>
        <button class="create-btn" onclick="window.location.href='register.php'">Create an Account</button>
        <?php if (isset($error)) { echo "<p class='error'>$error</p>"; } ?>
    </div>
</body>
</html>