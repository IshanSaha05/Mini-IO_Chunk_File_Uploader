package models

func (redisClient *RedisClient) ShutDown() {
	redisClient.Cancel()

	redisClient.Wg.Wait()

	redisClient.ExpireChannel.Close()
}
