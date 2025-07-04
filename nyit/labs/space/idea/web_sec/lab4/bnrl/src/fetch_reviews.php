<?php
session_start();
include 'db.php';

if (!isset($_GET['product_id'])) {
    echo json_encode([]);
    exit();
}

$product_id = intval($_GET['product_id']);
$query = "SELECT r.id, r.review_text, r.user_id, u.username 
          FROM reviews r 
          JOIN users u ON r.user_id = u.id 
          WHERE r.product_id = ?";
$stmt = $conn->prepare($query);
$stmt->bind_param("i", $product_id);
$stmt->execute();
$result = $stmt->get_result();

$reviews = [];
while ($row = $result->fetch_assoc()) {
        $row['is_owner'] = true; 
        $reviews[] = $row;
}
var_dump($reviews);
// ✅ Output JSON for frontend
$json = json_encode($reviews);
if ($json === false) {
    echo json_encode(["error" => json_last_error_msg()]);
} else {
    echo $json;
}

?>
