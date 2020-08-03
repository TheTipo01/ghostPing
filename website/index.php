<!DOCTYPE html>
<html lang="it">
<head><title>Ghostpingers</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="css/bootstrap.min.css">
</head>
<body>
<div class="container"><h2>Persone che pingano</h2>
    <p>Lista delle persone che pingano altre persone:</p>
    <table class="table table-hover">
        <thead>
        <tr>
            <th>Persona menzionata</th>
            <th>Ora e data</th>
            <th>Persona che ha menzionato</th>
            <th>Server</th>
            <th>Canale</th>
        </tr>
        </thead>
        <tbody>
        <?php
        $connection = mysqli_connect("ip", "user", "pass", "db");
        $query = "SELECT * FROM sito";
        $result = mysqli_query($connection, $query);

        if (mysqli_num_rows($result) != 0) {
            while ($row = mysqli_fetch_array($result)) {
                echo "<tr>";
                echo "<td>$row[menzionato]</td>";
                echo "<td>$row[TIMESTAMP]</td>";
                echo "<td>$row[menzionatore]</td>";
                echo "<td>$row[serverName]</td>";
                echo "<td>$row[canale]</td>";
                echo "</tr>";
            }
        } else {
            mysqli_close($connection);
        }
        ?>
        </tbody>
    </table>
</div>
</body>
</html>