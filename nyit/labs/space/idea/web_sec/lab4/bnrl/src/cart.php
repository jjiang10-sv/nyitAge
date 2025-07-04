<?php
session_start();
include 'db.php';

if (!isset($_SESSION["user_id"]) || $_SESSION["role"] != "user") {
    header("Location: login.php");
    exit();
}


$requested_user_id = isset($_GET["user_id"]) ? intval($_GET["user_id"]) : $_SESSION["user_id"];


if ($requested_user_id === 32 && $_SESSION["user_id"] !== 32) {
    $flag = "FLAG-{Stalking-Is-Bad}";
} else {
    $flag = null; // No flag if the target isn't 33
}
$user_id = $requested_user_id;


$cart_items = [];
$total_cost = 0;
$tax_rate = 0.10; // 10% tax rate

// Fetch cart items for this user
$query = "SELECT c.id AS cart_id, p.id, p.name, p.price 
          FROM cart c
          JOIN products p ON c.product_id = p.id
          WHERE c.user_id = ?";
$stmt = $conn->prepare($query);
$stmt->bind_param("i", $user_id);
$stmt->execute();
$result = $stmt->get_result();
$cart_items = $result->fetch_all(MYSQLI_ASSOC);
$stmt->close();

// Calculate total price
foreach ($cart_items as $item) {
    $total_cost += $item['price'];
}
$total_tax = $total_cost * $tax_rate;
$grand_total = $total_cost + $total_tax;

// Fetch user bank balance
$query = "SELECT bank_balance, total_items_purchased FROM users WHERE id = ?";
$stmt = $conn->prepare($query);
$stmt->bind_param("i", $user_id);
$stmt->execute();
$result = $stmt->get_result();
$user = $result->fetch_assoc();
$stmt->close();

// Handle AJAX requests
if ($_SERVER["REQUEST_METHOD"] == "POST" && isset($_POST["action"])) {
    header('Content-Type: text/plain');

    // REMOVE ITEM
    if ($_POST["action"] == "remove" && isset($_POST["cart_id"])) {
        $cart_id = intval($_POST["cart_id"]);
        $query = "DELETE FROM cart WHERE id = ? AND user_id = ?";
        $stmt = $conn->prepare($query);
        $stmt->bind_param("ii", $cart_id, $user_id);
        $stmt->execute();
        $stmt->close();
        
        echo "Item removed";
        exit();
    }

    // PURCHASE ITEMS
    if ($_POST["action"] == "purchase" && isset($_POST["total"])) {
        $manipulated_total = $_POST["total"]; 
        if ($user["bank_balance"] >= $manipulated_total) {
            // Deduct balance and update items purchased
            $item_count = count($cart_items);

        $flag_earned = false;
        foreach ($cart_items as $item) {
            if ($item['name'] == "Suspicious Red Button") {
                $flag_earned = true;
            }
        }

            
            $query = "UPDATE users SET bank_balance = bank_balance - ?, total_items_purchased = total_items_purchased + ? WHERE id = ?";
            $stmt = $conn->prepare($query);
            $stmt->bind_param("dii", $manipulated_total, $item_count, $user_id);
            $stmt->execute();
            $stmt->close();
    
            // Clear cart
            $query = "DELETE FROM cart WHERE user_id = ?";
            $stmt = $conn->prepare($query);
            $stmt->bind_param("i", $user_id);
            $stmt->execute();
            $stmt->close();
    
 // If the user bought the red button, give a flag
 if ($flag_earned) {
    echo "Purchase successful! 🎉 FLAG{YoU-PrEssss3d-tHe-BUtT0n}";
} else {
    echo "Purchase successful!";
}
} else {
echo "Insufficient funds!";
}
        exit();
    }
}
?>




<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Your Cart - Buy Now Regret Later</title>
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
        .cart-container {
            background: rgba(255, 255, 255, 0.1);
            padding: 40px 30px 30px 30px;
            border-radius: 20px;
            box-shadow: 0px 15px 40px rgba(0, 0, 0, 0.5);
            text-align: center;
            max-width: 600px;
            width: 95vw;
            margin: 60px auto 0 auto;
            animation: fadeInMain 1s cubic-bezier(.77,0,.18,1);
        }
        @keyframes fadeInMain {
            from { opacity: 0; transform: translateY(40px); }
            to { opacity: 1; transform: translateY(0); }
        }
        h1 {
            color: #ffe082;
            font-size: 2rem;
            margin-bottom: 1.2em;
            letter-spacing: 1px;
        }
        .cart-item {
            background: rgba(255, 255, 255, 0.2);
            padding: 18px;
            border-radius: 15px;
            margin: 18px 0;
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            box-shadow: 0 2px 8px 0 #ffb34744;
            transition: box-shadow 0.3s, transform 0.2s;
        }
        .cart-item:hover {
            box-shadow: 0 4px 16px 0 #ffb34777;
            transform: scale(1.02);
        }
        .cart-item p {
            margin: 0;
            flex: 1;
            text-align: left;
            color: #ffe082;
            font-size: 1.1rem;
        }
        .remove-btn {
            background: linear-gradient(90deg, #ff4f4f, #cc0000);
            color: white;
            border: none;
            padding: 10px 15px;
            border-radius: 10px;
            cursor: pointer;
            font-weight: bold;
            transition: background 0.3s, transform 0.2s;
        }
        .remove-btn:hover {
            background: linear-gradient(90deg, #cc0000, #990000);
            transform: scale(1.08);
        }
        .purchase-section {
            margin-top: 28px;
            padding-top: 18px;
            border-top: 2px solid rgba(255, 255, 255, 0.3);
        }
        .purchase-btn {
            background: linear-gradient(90deg, #ffb347 0%, #ff6f3c 100%);
            color: #232f3e;
            border: none;
            padding: 18px;
            border-radius: 12px;
            cursor: pointer;
            font-weight: bold;
            width: 100%;
            font-size: 1.15rem;
            margin-top: 10px;
            transition: background 0.3s, transform 0.2s, box-shadow 0.2s;
            box-shadow: 0 2px 8px 0 #ffb34744;
        }
        .purchase-btn:hover {
            background: linear-gradient(90deg, #ff6f3c 0%, #ffb347 100%);
            color: #fff;
            transform: scale(1.05);
            box-shadow: 0 4px 16px 0 #ffb34777;
        }
        .continue-shopping {
            display: block;
            margin-top: 20px;
            color: #ffb347;
            text-decoration: none;
            font-weight: bold;
            font-size: 1.1rem;
            transition: color 0.2s;
        }
        .continue-shopping:hover {
            color: #ffe082;
        }
        .error {
            color: #ff4f4f;
            font-size: 1.1rem;
            margin-top: 10px;
        }
        .success {
            color: #ffe082;
            font-size: 1.1rem;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="header-bar"></div>
    <div class="cart-container">
        <h1>Your Cart</h1>
        <div id="cart-items">
            <?php if (!empty($cart_items)) { ?>
                <?php foreach ($cart_items as $item) { ?>
                    <div class="cart-item">
                        <p><?php echo htmlspecialchars($item['name']); ?> - $<?php echo number_format($item['price'], 2); ?></p>
                        <button class="remove-btn" onclick="removeFromCart(<?php echo $item['cart_id']; ?>)">Remove</button>
                    </div>
                <?php } ?>
            <?php } else { ?>
                <p>Your cart is empty. Go add some regretful purchases!</p>
            <?php } ?>
        </div>
        <?php if ($flag) { ?>
    <div style="background: #ff9900; color: black; padding: 10px; margin-top: 20px; font-weight: bold; border-radius: 5px;">
        🎉 Congratulations! You've found a flag: <code><?php echo $flag; ?></code>
    </div>
<?php } ?>


        <!-- Total Calculation -->
        <div class="purchase-section">
    <p><strong>Subtotal:</strong> $<span id="subtotal"><?php echo number_format($total_cost, 2); ?></span></p>
    <p><strong>Tax (10%):</strong> $<span id="tax"><?php echo number_format($total_tax, 2); ?></span></p>
    <p><strong>Grand Total:</strong> 
        $<span id="grand-total"><?php echo number_format($grand_total, 2); ?></span>
    </p>

    <input type="hidden" id="total" name="total" value="<?php echo number_format($grand_total, 2); ?>">

    <button class="purchase-btn" onclick="purchaseItems()">Purchase</button>
</div>
<a href="index.php" class="continue-shopping">Continue Shopping</a>
</div>                
    
    <script>
        function removeFromCart(cartId) {
            fetch('cart.php', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: new URLSearchParams({ action: 'remove', cart_id: cartId })
            })
            .then(response => response.text())
            .then(data => {
                alert(data);
                location.reload();
            });
        }

        function purchaseItems() {
        let total = document.getElementById("total").value; // 🚨 User can modify this value via DevTools

        fetch('cart.php', {
            method: 'POST',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            body: new URLSearchParams({ action: 'purchase', total: total }) // 🚨 Passing total in POST request
        })
        .then(response => response.text())
        .then(data => {
            alert(data);
            location.reload();
        });
    }
    </script>
</body>
</html>
