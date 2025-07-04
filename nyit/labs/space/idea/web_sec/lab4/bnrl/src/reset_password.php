<?php
session_start();
include 'db.php';

if ($_SERVER["REQUEST_METHOD"] == "POST") {
    $new_password = password_hash($_POST["password"], PASSWORD_BCRYPT);
    $email = $_SESSION["reset_email"];

    $query = "UPDATE users SET password = ? WHERE email = ?";
    $stmt = $conn->prepare($query);
    $stmt->bind_param("ss", $new_password, $email);
    $stmt->execute();

    session_destroy();
    echo "<script>alert('✅ Password Reset Successfully!'); window.location.href='login.php';</script>";
}
?>

<!DOCTYPE html>
<html>
<head>
    <title>Reset Password</title>
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
            padding: 40px 30px 30px 30px;
            border-radius: 20px;
            box-shadow: 0px 15px 40px rgba(0, 0, 0, 0.5);
            text-align: center;
            max-width: 400px;
            width: 95vw;
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
        input[type="password"] {
            width: 95%;
            padding: 12px;
            border-radius: 10px;
            border: none;
            font-size: 16px;
            background: rgba(255, 255, 255, 0.2);
            color: white;
            text-align: center;
            margin-bottom: 18px;
            transition: background 0.3s, transform 0.2s;
        }
        input[type="password"]:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: scale(1.02);
        }
        .submit-btn {
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
        .submit-btn:hover {
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
        <h2>Reset Password</h2>
        <form method="post">
            <label>New Password:</label>
            <input type="password" name="password" placeholder="Create a new password" required>
            <button type="submit" class="submit-btn">Reset</button>
        </form>
    </div>
</body>
</html>
