EXECUTE_TIME=$(date -v+2M +"%Y-%m-%d %H:%M:%S")

echo "Запланируем запись в очередь на: $EXECUTE_TIME"
echo "Параметры:"
echo "   - Queue ID: 12"
echo "   - Group Code: TEST_GROUP"
echo "   - Slot Time: 10:00"
echo ""


RESPONSE=$(curl -s -X POST http://localhost:8090/api/jobs/join-queue \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"queueId\": 1,
    \"groupCode\": \"1\",
    \"slotTime\": \"10:00\",
    \"executeAt\": \"$EXECUTE_TIME\"
  }")

echo "Ответ:"
echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

if echo "$RESPONSE" | grep -q '"status":"success"'; then
    echo "Запись в очередь запланирована"
    echo ""
    echo "Задача выполнится автоматически через 2 минуты"
else
    echo "Ошибка при создании задачи"
fi

