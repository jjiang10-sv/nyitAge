<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Vulnerable Web App</title>
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
        input[type="text"], input[type="email"], input[type="password"] {
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
        input[type="text"]:hover, input[type="email"]:hover, input[type="password"]:hover {
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
        .back-btn {
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
        .back-btn:hover {
            background: linear-gradient(90deg, #ff6f3c 0%, #ffb347 100%);
            color: #fff;
            transform: scale(1.05);
            box-shadow: 0 4px 16px 0 #ffb34777;
        }
    </style>
    <script>
        document.addEventListener("DOMContentLoaded", function() {
            document.getElementById("registerForm").addEventListener("submit", function(event) {
                event.preventDefault();
                showSuccessMessage();
            });
        });

        function showSuccessMessage() {
            var popup = document.createElement("div");
            popup.className = "popup";
            popup.innerHTML = "<p>🎉 Registration successful! You can now <a href='login.php' style='color: #fff; text-decoration: underline;'>login</a>.</p><button onclick='closePopup()'>OK</button>";
            document.body.appendChild(popup);
            popup.style.display = "block";
        }

        function closePopup() {
            var popup = document.querySelector(".popup");
            if (popup) {
                popup.style.display = "none";
                document.getElementById("registerForm").submit();
            }
        }
    </script>
</head>
<body>
    <div class="header-bar"></div>
    <div class="container">
        <h2>Register</h2>
        <form id="registerForm" action="register.php" method="post">
            <label>Username:</label>
            <input type="text" name="username" placeholder="Enter your username" required>
            
            <label>Email:</label>
            <input type="email" name="email" placeholder="Enter your email" required>
            
            <label>Password:</label>
            <input type="password" name="password" placeholder="Create a password" required>
            
            <input type="submit" value="Register">
        </form>
        <button class="back-btn" onclick="window.location.href='login.php'">Back to Login</button>
    </div>
</body>
</html>



<?php
include 'db.php'; // Include database connection

if ($_SERVER["REQUEST_METHOD"] == "POST") {
    $username = $_POST["username"];
    $email = $_POST["email"];
    $password = password_hash($_POST["password"], PASSWORD_BCRYPT); // Hash password

    // Insert into database
    $query = "INSERT INTO users (username, email, password) VALUES (?, ?, ?)";
    $stmt = $conn->prepare($query);
    $stmt->bind_param("sss", $username, $email, $password);

    if ($stmt->execute()) {
        echo "<script>alert('Registration successful!'); window.location.href='login.php';</script>";
    } else {
        echo "<script>alert('Error: Could not register user. " . $stmt->error . "');</script>";
    }

    $stmt->close();
    $conn->close();
}
?>
