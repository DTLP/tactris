import pygame
import random

# Define constants
BLACK = (0, 0, 0)
WHITE = (255, 255, 255)
GREY = (100, 100, 150)
WIDTH = 40
HEIGHT = 40
MARGIN = 1

# Define functions
def generate_block():
    shapes = [
        [[0, 0], [1, 0], [2, 0], [3, 0]],  # I shape
		[[0, 0], [0, 1], [0, 2], [0, 3]],  # I horizontal shape
        [[0, 0], [0, 1], [0, 2], [1, 2]],  # L shape
		[[0, 0], [1, 0], [2, 0], [0, 1]],  # L horizontal shape
        [[0, 0], [0, 1], [0, 2], [1, 0]],  # J shape
		[[0, 0], [1, 0], [2, 0], [2, 1]],  # J horizontal shape
        [[0, 0], [0, 1], [1, 1], [1, 0]],  # O shape
        [[0, 0], [0, 1], [1, 1], [1, 2]],  # S shape
        [[0, 1], [0, 2], [1, 0], [1, 1]],  # Z shape
		[[0, 0], [0, 1], [0, 2], [1, 1]],  # T shape
    ]
    return random.choice(shapes)

def can_place_block(block, position, grid):
    for b in block:
        row = position[0] + b[0]
        col = position[1] + b[1]
        if row < 0 or row >= 10 or col < 0 or col >= 10:
            return False
        if grid[row][col] == 1:
            return False
    return True

def add_block(block, position, grid):
    for b in block:
        row = position[0] + b[0]
        col = position[1] + b[1]
        grid[row][col] = 1

def clear_lines(grid):
    full_rows = [row for row in range(len(grid)) if all(grid[row][col] != (255, 255, 255) for col in range(len(grid[row])))]
    #full_rows = [row for row in range(len(grid)) if all(grid[row][col] == 0 for col in range(len(grid[row])))]
    full_cols = [col for col in range(len(grid[0])) if all(grid[row][col] != (255, 255, 255) for row in range(len(grid)))]
    #full_cols = [col for col in range(len(grid[0])) if all(grid[row][col] == 0 for row in range(len(grid)))]

    for row in full_rows:
        grid.pop(row)
        grid.insert(0, [(255, 255, 255)] * len(grid[0]))

    for col in full_cols:
        for row in range(len(grid)):
            del grid[row][col]
        for row in range(len(grid)):
            grid[row].insert(0, (255, 255, 255))

def reset_game():
    global score, grid

    print("Game over. Score: " + str(score))
	
    current_block = generate_block()
    next_block = generate_block()
    score = 0
    grid = []
    for row in range(10):
        grid.append([0] * 10)
		
    # Update the screen
    pygame.display.flip()

# Initialize Pygame
pygame.init()

# Set the height and width of the screen
WINDOW_SIZE = [(MARGIN + WIDTH) * 20 + MARGIN, (MARGIN + HEIGHT) * 10 + MARGIN]
screen = pygame.display.set_mode(WINDOW_SIZE)

# Set the title of the window
pygame.display.set_caption("Tactris")

# Set the font for the score
font = pygame.font.Font(None, 36)

# Initialize the grid
grid = []
for row in range(10):
    grid.append([0] * 10)

# Initialize the current block and position
current_block = generate_block()
next_block = generate_block()
current_position = [1, 10]
next_position = [1, 15]

# Initialize the score
score = 0

# set up the game reset button
button_width = 140
button_height = 60
button_x = WINDOW_SIZE[0] - 200
button_y = WINDOW_SIZE[1] - 80
button_color = (0, 255, 0)
button_text = "Reset"
font_size = 24
font = pygame.font.SysFont(None, font_size)
text_color = (255, 255, 255)
button_rect = pygame.Rect(button_x, button_y, button_width, button_height)

# Loop until the user clicks the close button
done = False

# Set the clock
clock = pygame.time.Clock()

# Loop until the game is over
while not done:

    # Handle events
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            done = True
        elif event.type == pygame.MOUSEBUTTONDOWN and event.button == 1:
            mouse_position = pygame.mouse.get_pos()

            # Check if the reset button was clicked
            if button_rect.collidepoint(event.pos):
                reset_game()

            # Calculate the grid position of the mouse click
            row = mouse_position[1] // (HEIGHT + MARGIN)
            col = mouse_position[0] // (WIDTH + MARGIN)

            # Check if the current block can be placed at the clicked position
            if can_place_block(current_block, [row, col], grid):
                add_block(current_block, [row, col], grid)
                clear_lines(grid)
                current_block = next_block
                next_block = generate_block()
                current_position = [1, 10]
                next_position = [1, 15]
                score += 1

    # Draw the grid
    screen.fill(BLACK)
    for row in range(10):
        for col in range(10):
            color = WHITE
            if grid[row][col] == 1:
                color = GREY
            pygame.draw.rect(screen, color, [(MARGIN + WIDTH) * col + MARGIN, (MARGIN + HEIGHT) * row + MARGIN, WIDTH, HEIGHT])

    # Draw the current block
    cur_text = font.render("Current: ", True, WHITE)
    screen.blit(cur_text, [WINDOW_SIZE[0] - 400, WINDOW_SIZE[1] - 400])
	
    for b in current_block:
	    row = current_position[0] + b[0]
	    col = current_position[1] + b[1]
	    pygame.draw.rect(screen, WHITE, [(MARGIN + WIDTH) * col + MARGIN + 5, (MARGIN + HEIGHT) * row + MARGIN, WIDTH, HEIGHT])
		
    # Draw the next block
    next_text = font.render("Next: " , True, WHITE)
    screen.blit(next_text, [WINDOW_SIZE[0] - 200, WINDOW_SIZE[1] - 400])
	
    for b in next_block:
	    row = next_position[0] + b[0]
	    col = next_position[1] + b[1]
	    pygame.draw.rect(screen, WHITE, [(MARGIN + WIDTH) * col + MARGIN + 5, (MARGIN + HEIGHT) * row + MARGIN, WIDTH, HEIGHT])

    # Draw the score
    score_text = font.render("Score: " + str(score), True, WHITE)
    screen.blit(score_text, [WINDOW_SIZE[0] - 400, WINDOW_SIZE[1] - 120])

    # Draw the game reset button
    pygame.draw.rect(screen, button_color, button_rect)
	
    pygame.draw.rect(screen, BLACK, button_rect.inflate(10, 10))
    pygame.draw.rect(screen, WHITE, button_rect, 2)

    text_surface = font.render(button_text, True, text_color)
    text_rect = text_surface.get_rect(center=button_rect.center)
    reset_text = font.render("RESET", True, WHITE)
    reset_button = screen.blit(reset_text, text_rect)

    # Update the screen
    pygame.display.flip()

    # Wait for the next frame
    clock.tick(60)

# Quit Pygame
pygame.quit()
